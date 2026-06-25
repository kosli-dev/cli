package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// serviceAccount models a single service account as returned by the
// create/get/update endpoints (ServiceAccountResponse) and, as a subset, by
// the list endpoint (ServiceAccountListItemResponse). Timestamps are epoch
// seconds (float64), matching the rest of the v2 API (see apiKey.go).
//
// NOTE: the exact field set is mirrored from the server's ServiceAccount*
// response models. If the API adds/renames fields, update this struct and the
// fixtures under testdata/service-account.
type serviceAccount struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Privilege   string  `json:"privilege"`
	CreatedAt   float64 `json:"created_at"`
}

// printServiceAccountAsTable renders a single service account (the create/get/
// update response) as a key:value table.
func printServiceAccountAsTable(raw string, out io.Writer, page int) error {
	var sa serviceAccount
	if err := json.Unmarshal([]byte(raw), &sa); err != nil {
		return err
	}

	rows, err := serviceAccountTableRows(sa)
	if err != nil {
		return err
	}
	tabFormattedPrint(out, []string{}, rows)
	return nil
}

// printServiceAccountsListAsTable renders the list response as a table.
func printServiceAccountsListAsTable(raw string, out io.Writer, page int) error {
	var accounts []serviceAccount
	if err := json.Unmarshal([]byte(raw), &accounts); err != nil {
		return err
	}

	if len(accounts) == 0 {
		logger.Info("No service accounts were found.")
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "PRIVILEGE", "CREATED"}
	rows := []string{}
	for _, sa := range accounts {
		createdAt, err := optionalTimestamp(sa.CreatedAt)
		if err != nil {
			return err
		}
		description := sa.Description
		if description == "" {
			description = "N/A"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s", sa.Name, description, sa.Privilege, createdAt))
	}
	tabFormattedPrint(out, header, rows)
	return nil
}

// serviceAccountTableRows builds the key:value rows describing one service account.
func serviceAccountTableRows(sa serviceAccount) ([]string, error) {
	createdAt, err := optionalTimestamp(sa.CreatedAt)
	if err != nil {
		return nil, err
	}
	description := sa.Description
	if description == "" {
		description = "N/A"
	}

	rows := []string{}
	rows = append(rows, fmt.Sprintf("Name:\t%s", sa.Name))
	rows = append(rows, fmt.Sprintf("Description:\t%s", description))
	rows = append(rows, fmt.Sprintf("Privilege:\t%s", sa.Privilege))
	rows = append(rows, fmt.Sprintf("Created At:\t%s", createdAt))
	return rows, nil
}
