package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/hashing"
	"collector/pkg/retry"
)

type SendMetricResult struct {
	Code   int
	Status string
	Body   string
}

func NewSendMetricResult(response *http.Response) *SendMetricResult {
	body, _ := io.ReadAll(response.Body)
	return &SendMetricResult{
		Code:   response.StatusCode,
		Status: response.Status,
		Body:   string(body),
	}
}

func worker(client *http.Client, jobs <-chan *http.Request, results chan<- *SendMetricResult) {
	for request := range jobs {
		var sendMetricResult *SendMetricResult
		_ = retry.Try(func() error {
			var err error
			sendMetricResult, err = sendRequest(client, request)
			return err
		})

		results <- sendMetricResult
	}
}

func sendData(
	ctx context.Context,
	client *http.Client,
	conf *config.AgentConfig,
	stats []*domain.MetricForm,
) error {
	url := conf.GetAddress()

	poolSize := len(stats)

	jobs := make(chan *http.Request, poolSize)
	results := make(chan *SendMetricResult, poolSize)

	workers := conf.RateLimit
	for w := 1; w <= workers; w++ {
		go worker(client, jobs, results)
	}

	for _, form := range stats {
		var endpoint string
		if form.IsGaugeType() {
			endpoint = fmt.Sprintf("http://%s/update/gauge/%s/%f", url, form.ID, *form.Value)
		} else if form.IsCounterType() {
			endpoint = fmt.Sprintf("http://%s/update/counter/%s/%d", url, form.ID, *form.Delta)
		} else {
			return fmt.Errorf("invalid metric type: %v", form.MType)
		}

		data, marshErr := json.Marshal(form)
		if marshErr != nil {
			return fmt.Errorf("marshall data error: %w", marshErr)
		}

		hash := conf.GetHashKey()
		if hash != "" {
			hash = hashing.HashByKey(string(data), conf.GetHashKey())
		}

		req, reqErr := buildRequest(ctx, endpoint, data, hash)
		if reqErr != nil {
			return fmt.Errorf("build request error: %w", reqErr)
		}

		jobs <- req
	}

	close(jobs)

	return nil
}

func sendRequest(client *http.Client, req *http.Request) (*SendMetricResult, error) {
	defer req.Body.Close()

	resp, respErr := client.Do(req)

	if respErr != nil {
		return nil, fmt.Errorf("response error: %w", respErr)
	}

	defer resp.Body.Close()

	_, bodyErr := io.Copy(io.Discard, resp.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf("read body error: %w", bodyErr)
	}

	return NewSendMetricResult(resp), nil
}

func buildRequest(
	ctx context.Context,
	url string,
	data []byte,
	hash string,
) (*http.Request, error) {
	req, reqErr := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(data),
	)

	if reqErr != nil {
		return nil, fmt.Errorf("request error: %w", reqErr)
	}

	req.Header.Add("Content-Type", "application/json")

	if hash != "" {
		req.Header.Add(domain.HashHeader, hash)
	}

	return req, nil
}
