package collector

//
// option.go
// Copyright (C) 2020 gaspar_d </var/spool/mail/gaspar_d>
//
// Distributed under terms of the MIT license.
//

import "io"
import "net/http"

// MustGetNewRequest creates a new HTTP Request and set all
// the required headers to identify the ccloudexporter
func MustGetNewRequest(method string, endpoint string, reader io.Reader) *http.Request {
	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		panic(err)
	}

	apikey := MustGetAPIKey()
	apisecret := MustGetAPISecret()

	req.SetBasicAuth(apikey, apisecret)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "ccloudexporter/"+Version)
	req.Header.Add("Correlation-Context", "service.name=ccloudexporter,service.version="+Version)

	return req
}
