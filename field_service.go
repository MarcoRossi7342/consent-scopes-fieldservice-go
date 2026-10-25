package main

import (
	"context"
	"errors"
)

type WorkOrder struct {
	ID             string `json:"id"`
	PhotoReference string `json:"photo_reference"`
	DispatchStatus string `json:"dispatch_status"`
	FollowUp       string `json:"follow_up"`
}

var ErrConsentRequired = errors.New("work-order photo consent is required")

func LoadWorkOrder(ctx context.Context, client *ConsentClient, userID string) (WorkOrder, error) {
	status, err := client.Check(ctx, userID, "work_order_photos")
	if err != nil {
		return WorkOrder{}, err
	}
	if !status.Result {
		return WorkOrder{}, ErrConsentRequired
	}
	return WorkOrder{ID: "wo-1042", PhotoReference: "photo-1042", DispatchStatus: "en_route", FollowUp: "confirm repair notes after visit"}, nil
}
