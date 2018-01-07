package client

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alpacahq/marketstore/frontend"
	"github.com/alpacahq/marketstore/utils/io"
	"github.com/alpacahq/marketstore/utils/rpc/msgpack2"
)

type Client struct {
	BaseURL string
}

func NewClient(baseurl string) (cl *Client, err error) {
	cl = new(Client)
	_, err = url.Parse(baseurl)
	if err != nil {
		return nil, err
	}
	cl.BaseURL = baseurl
	return cl, nil
}

func (cl *Client) DoRPC(functionName string, args interface{}) (csm io.ColumnSeriesMap, err error) {
	/*
		Does a remote procedure call using the msgpack2 protocol for RPC that return a QueryReply
	*/
	if args == nil {
		return nil, fmt.Errorf("args must be non-nil - have: args: %v\n",
			args)
	}
	message, err := msgpack2.EncodeClientRequest("DataService."+functionName, args)
	if err != nil {
		return nil, err
	}
	reqURL := cl.BaseURL + "rpc"
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(message))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-msgpack")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch functionName {
	case "Query", "SQLStatement":
		result := &frontend.MultiQueryResponse{}
		err = msgpack2.DecodeClientResponse(resp.Body, result)
		if err != nil {
			fmt.Printf("Error decoding: %s\n", err)
			return nil, err
		}
		return ConvertMultiQueryReplyToColumnSeries(result)
	case "FeedStart":
		result := &frontend.FeedStartReply{}
		err = msgpack2.DecodeClientResponse(resp.Body, result)
		if len(result.ErrorText) != 0 {
			return nil, fmt.Errorf("%s", result.ErrorText)
		}
		csm := io.NewColumnSeriesMap()
		cs := io.NewColumnSeries()
		cs.AddColumn("PID", []int{result.PID})
		tbk, _ := io.NewTimeBucketKeyFromString("NA:NA")
		csm.AddColumnSeries(*tbk, cs)
		return csm, nil
	case "FeedList":
		result := &frontend.FeedListReply{}
		err = msgpack2.DecodeClientResponse(resp.Body, result)
		csm := io.NewColumnSeriesMap()
		cs := io.NewColumnSeries()
		cs.AddColumn("Descriptions", result.Descriptions)
		tbk, _ := io.NewTimeBucketKeyFromString("NA:NA")
		csm.AddColumnSeries(*tbk, cs)
		return csm, nil
	case "FeedKill":
		result := &frontend.FeedKillReply{}
		err = msgpack2.DecodeClientResponse(resp.Body, result)
		return nil, nil
	case "Write":
		result := &frontend.MultiWriteResponse{}
		err = msgpack2.DecodeClientResponse(resp.Body, result)
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported RPC response")
	}
}

func ConvertMultiQueryReplyToColumnSeries(result *frontend.MultiQueryResponse) (csm io.ColumnSeriesMap, err error) {
	if result == nil {
		return nil, nil
	}
	csm = io.NewColumnSeriesMap()
	for _, ds := range result.Responses { // Datasets are packed in a slice, each has a NumpyMultiDataset inside
		nmds := ds.Result
		for tbkStr, startIndex := range nmds.StartIndex {
			cs, err := nmds.ToColumnSeries(startIndex, nmds.Lengths[tbkStr])
			if err != nil {
				return nil, err
			}
			tbk, _ := io.NewTimeBucketKeyFromString(tbkStr)
			csm[*tbk] = cs
		}
	}
	return csm, nil
}

func ColumnSeriesFromResult(shapes []io.DataShape, columns map[string]interface{}) (cs *io.ColumnSeries, err error) {
	cs = io.NewColumnSeries()
	for _, shape := range shapes {
		name := shape.Name
		typ := shape.Type
		base := columns[name].([]interface{})
		if base == nil {
			return nil, fmt.Errorf("unable to unpack %s", name)
		}
		i_column, err := io.CreateSliceFromSliceOfInterface(base, typ)
		if err != nil {
			return nil, err
		}
		cs.AddColumn(name, i_column)
	}
	return cs, nil
}
