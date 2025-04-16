package vigilant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	registrationEndpoint = "/api/registration"
	heartbeatEndpoint    = "/api/registration/heartbeat"
)

type registrationHandler struct {
	token    string
	endpoint string
	client   *http.Client

	registered            bool
	serviceName           string
	serviceInstanceNumber int
	serviceInstanceId     uuid.UUID

	doneChan       chan struct{}
	registeredChan chan struct{}
	wg             sync.WaitGroup
	mux            sync.RWMutex
}

func newRegistrationHandler(
	token string,
	endpoint string,
	serviceName string,
	client *http.Client,
) *registrationHandler {
	return &registrationHandler{
		token:          token,
		endpoint:       endpoint,
		serviceName:    serviceName,
		client:         client,
		registeredChan: make(chan struct{}),
		doneChan:       make(chan struct{}),
		wg:             sync.WaitGroup{},
		mux:            sync.RWMutex{},
	}
}

func (h *registrationHandler) start() {
	log.Printf("Starting registration handler for service %s", h.serviceName)
	h.wg.Add(1)
	go h.runRegistration()
}

func (h *registrationHandler) stop() {
	log.Printf("Stopping registration handler for service %s", h.serviceName)
	close(h.doneChan)
	h.wg.Wait()
	h.deregister()
	log.Printf("Registration handler for service %s stopped", h.serviceName)
}

func (h *registrationHandler) getServiceInstance() (string, error) {
	log.Printf("Getting service instance for service %s", h.serviceName)
	h.mux.RLock()
	defer h.mux.RUnlock()

	if !h.registered {
		return "", errors.New("not registered")
	}

	return fmt.Sprintf("%s-%d", h.serviceName, h.serviceInstanceNumber), nil
}

func (h *registrationHandler) waitForRegistration(ctx context.Context) error {
	log.Printf("Waiting for registration for service %s", h.serviceName)
	h.mux.RLock()
	registered := h.registered
	h.mux.RUnlock()
	if registered {
		log.Printf("Registration for service %s already completed", h.serviceName)
		return nil
	}

	select {
	case <-h.registeredChan:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *registrationHandler) runRegistration() {
	log.Printf("Running registration for service %s", h.serviceName)
	defer h.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !h.registered {
				log.Printf("Registering service %s", h.serviceName)
				var err error
				for range [10]int{} {
					err = h.register()
					if err == nil {
						log.Printf("Registration for service %s completed", h.serviceName)
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
				if err != nil {
					log.Printf("Error registering service %s: %v", h.serviceName, err)
					return
				}
			} else {
				log.Printf("Heartbeating service %s", h.serviceName)
				h.heartbeat()
			}
		case <-h.doneChan:
			log.Printf("Stopping registration handler for service %s", h.serviceName)
			return
		}
	}
}

func (h *registrationHandler) register() error {
	response, err := h.sendRegistrationRequest()
	if err != nil {
		return err
	}

	h.mux.Lock()
	defer h.mux.Unlock()
	if !h.registered {
		h.serviceInstanceNumber = response.ServiceInstanceNumber
		h.serviceInstanceId = uuid.MustParse(response.ServiceInstanceID)
		h.registered = true
		close(h.registeredChan)
	}

	return nil
}

func (h *registrationHandler) deregister() {
	err := h.sendDeregistrationRequest()
	if err != nil {
		return
	}

	h.mux.Lock()
	defer h.mux.Unlock()
	h.registered = false
	h.serviceInstanceNumber = 0
	h.serviceInstanceId = uuid.Nil
}

func (h *registrationHandler) heartbeat() {
	response, err := h.sendHeartbeatRequest()
	if err != nil {
		return
	}

	h.mux.Lock()
	defer h.mux.Unlock()

	if !response.Reassigned {
		return
	}

	h.serviceInstanceNumber = response.NewInstanceNumber
	h.serviceInstanceId = response.NewInstanceID
	h.registered = true
}

func (h *registrationHandler) sendRegistrationRequest() (*registrationResponse, error) {
	request := &registrationRequest{
		Token:       h.token,
		ServiceName: h.serviceName,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", h.endpoint+registrationEndpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response registrationResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (h *registrationHandler) sendDeregistrationRequest() error {
	request := &deregistrationRequest{
		Token:                 h.token,
		ServiceName:           h.serviceName,
		ServiceInstanceNumber: h.serviceInstanceNumber,
		ServiceInstanceID:     h.serviceInstanceId,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", h.endpoint+registrationEndpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (h *registrationHandler) sendHeartbeatRequest() (*heartbeatResponse, error) {
	request := &heartbeatRequest{
		Token:                 h.token,
		ServiceName:           h.serviceName,
		ServiceInstanceNumber: h.serviceInstanceNumber,
		ServiceInstanceID:     h.serviceInstanceId,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", h.endpoint+heartbeatEndpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response heartbeatResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type registrationRequest struct {
	Token       string `json:"token"`
	ServiceName string `json:"service_name"`
}

type registrationResponse struct {
	ServiceInstanceNumber int    `json:"service_instance_number"`
	ServiceInstanceID     string `json:"service_instance_id"`
}

type deregistrationRequest struct {
	Token                 string    `json:"token"`
	ServiceName           string    `json:"service_name"`
	ServiceInstanceNumber int       `json:"service_instance_number"`
	ServiceInstanceID     uuid.UUID `json:"service_instance_id"`
}

type heartbeatRequest struct {
	Token                 string    `json:"token"`
	ServiceName           string    `json:"service_name"`
	ServiceInstanceNumber int       `json:"service_instance_number"`
	ServiceInstanceID     uuid.UUID `json:"service_instance_id"`
}

type heartbeatResponse struct {
	Reassigned        bool      `json:"reassigned"`
	NewInstanceNumber int       `json:"new_instance_number"`
	NewInstanceID     uuid.UUID `json:"new_instance_id"`
}
