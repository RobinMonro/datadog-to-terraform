package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rodaine/hclencoder"

	"github.com/juliogreff/datadog-to-terraform/pkg/types"
)

const (
	ddUrl = "https://api.datadoghq.com"

	dashboardResource = "dashboard"
	monitorResource   = "monitor"
)

func request(method, url string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return client.Do(req)
}

func main() {
	body := `{
		"id": 0,
		"name": "",
		"type": "metric alert",
		"query": "avg(last_5m):max:aws.rds.replica_lag{*} by {dbinstanceidentifier} > 3600",
		"message": "",
		"tags": [],
		"options": {
			"notify_audit": false,
			"locked": false,
			"timeout_h": 0,
			"new_host_delay": 300,
			"require_full_window": true,
			"notify_no_data": false,
			"escalation_message": "",
			"no_data_timeframe": null,
			"include_tags": true,
			"thresholds": {}
		},
		"priority": null
	}`

	resourceType := "monitor"

	resourceId := "monitor_name"
	if resourceType == "dashboard" {
		resourceId = body[len(body)-13 : len(body)-2]
	}

	resource := types.Resource{Name: resourceId}

	switch resourceType {
	case dashboardResource:
		var dashboard *types.Board
		err := json.Unmarshal([]byte(body), &dashboard)
		if err != nil {
			fail("%s %s: unable to parse JSON: %s", resourceType, resourceId, err)
		}

		resource.Type = "datadog_dashboard"
		resource.Board = dashboard
	case monitorResource:
		var monitor *types.Monitor
		err := json.Unmarshal([]byte(body), &monitor)
		if err != nil {
			fail("%s %s: unable to parse JSON: %s", resourceType, resourceId, err)
		}

		resource.Type = "datadog_monitor"
		resource.Monitor = monitor
	}

	hcl, err := hclencoder.Encode(types.ResourceWrapper{
		Resource: resource,
	})
	if err != nil {
		fail("%s %s: unable to encode hcl: %s", resourceType, resourceId, err)
	}

	fmt.Println(string(hcl))
}

func fail(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
