package cloudflare

import (
	"encoding/json"
	"errors"
)

type DnsRecordsResponse struct {
	Success  bool             `json:"success"`
	Errors   []ResponseErrors `json:"errors"`
	Result   DnsRecords       `json:"result"`
	Messages []string         `json:"messages"`
}

type DnsRecords []DnsRecord

// UnmarshalJSON handles unmarshalling from either an array or a single object.
func (dr *DnsRecords) UnmarshalJSON(data []byte) error {
	// Try unmarshalling data into a slice first.
	var records []DnsRecord
	if err := json.Unmarshal(data, &records); err == nil {
		*dr = records
		return nil
	}

	// If that fails, try unmarshalling into a single record.
	var record DnsRecord
	if err := json.Unmarshal(data, &record); err == nil {
		*dr = []DnsRecord{record}
		return nil
	}

	return errors.New("DnsRecords: data is neither an array nor a single object")
}
