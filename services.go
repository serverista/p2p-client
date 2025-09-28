package p2pclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ServiceAction represents the management action on a service
type ServiceAction string

const (
	// ServiceStart starts a service
	ServiceStart = "START"
	// ServiceShutdown stops a service
	ServiceShutdown = "SHUTDOWN"
	// ServiceRestart restarts a service
	ServiceRestart = "RESTART"
	// ServiceReinstall reinstalls the service to its initial state
	ServiceReinstall = "REINSTALL"
	// ServiceDelete destroys the service
	ServiceDelete = "DELETE"
)

// CreateServiceRequest is the service creation payload.
type CreateServiceRequest struct {
	PlanID       uint   `json:"plan_id"`
	OS           Os     `json:"os"`
	Amount       int    `json:"amount"`
	SSHPublicKey string `json:"ssh_public_key"`
	Name         string `json:"name"`
}

// CreateServices creates a new service given an optional custom name
// plan id, os type, number of instances and a public key.
func (c *Client) CreateServices(ctx context.Context, request CreateServiceRequest, nonce string) ([]Service, error) {
	if request.PlanID == 0 {
		return nil, errors.New("plan id is required")
	}

	if request.OS == "" {
		return nil, errors.New("os is required")
	}

	if request.Amount <= 0 {
		return nil, errors.New("number of instances must be greater than zero")
	}

	if request.SSHPublicKey == "" {
		return nil, errors.New("ssh public key is required")
	}

	req := CreateServiceRequest{
		PlanID:       request.PlanID,
		OS:           request.OS,
		Amount:       request.Amount,
		SSHPublicKey: request.SSHPublicKey,
		Name:         request.Name,
	}

	bts, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create service request: %w", err)
	}

	resp, err := c.request(ctx, CreateServicesEndpoint.Method, CreateServicesEndpoint.Uri, bts, nonce, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	var services []Service
	err = json.Unmarshal(resp.Body, &services)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal CreateService: %w", err)
	}

	return services, nil
}

// ListServices returns a list of available services.
func (c *Client) ListServices(ctx context.Context, nonce string) ([]Service, error) {
	resp, err := c.request(ctx, ListUserServicesEndpoint.Method, ListUserServicesEndpoint.Uri, nil, nonce, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	var services []Service
	err = json.Unmarshal(resp.Body, &services)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ListServices: %w", err)
	}

	return services, nil
}

// Get a specific service.
func (c *Client) GetService(ctx context.Context, id uint, nonce string) (*Service, error) {
	resp, err := c.request(ctx, GetUserServiceEndpoint.Method, fmt.Sprintf(GetUserServiceEndpoint.Uri, id), nil, nonce, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	var service Service
	err = json.Unmarshal(resp.Body, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal GetService: %w", err)
	}

	return &service, nil
}

type serviceActionRequest struct {
	Action string `json:"action"`
}

type serviceActionResponse struct {
	Error    string `json:"error"`
	ActionID string `json:"action_id"`
}

// ServiceAction performs an action such as start, shutdown, restart, reinstall and delete on a service.
func (c *Client) ServiceAction(ctx context.Context, action ServiceAction, id uint, nonce string) error {
	if action == "" {
		return errors.New("action is required")
	}
	reqBody, err := json.Marshal(serviceActionRequest{
		Action: string(action),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal service action request: %w", err)
	}
	resp, err := c.request(ctx, ManageServiceEndpoint.Method, fmt.Sprintf(ManageServiceEndpoint.Uri, id), reqBody, nonce, time.Now().Unix())
	if err != nil {
		return err
	}

	var actionResp serviceActionResponse
	err = json.Unmarshal(resp.Body, &actionResp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ServiceAction: %w", err)
	}

	if actionResp.Error != "" {
		return errors.New(actionResp.Error)
	}

	if actionResp.ActionID == "" {
		return errors.New("failed to perform action")
	}

	return nil
}
