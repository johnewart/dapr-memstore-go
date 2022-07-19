package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/dapr/components-contrib/state"
	"github.com/dapr/components-contrib/state/utils"
	sdk "github.com/dapr/dapr/pkg/components/pluggable/sdk/state"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"google.golang.org/grpc"
	"zombiezen.com/go/log"
)

// MemoryStore is a store used for testing
type MemoryStore struct {
	ctx  context.Context
	data map[string][]byte
}

func NewMemoryStore(ctx context.Context) *MemoryStore {
	return &MemoryStore{
		data: map[string][]byte{},
		ctx:  ctx,
	}
}

func (s *MemoryStore) Serve(socket string) error {
	if _, err := os.Stat(socket); err == nil {
		if err := os.RemoveAll(socket); err != nil {
			log.Errorf(s.ctx, "Error removing socket: %s", err)
		}
	}

	srv := sdk.GRPCStateStoreServer{
		Impl: s,
	}

	originalUmask := syscall.Umask(0)
	syscall.Umask(0)

	if listener, err := net.Listen("unix", socket); err != nil {
		return fmt.Errorf("unable to open listener server socket %s: %v", socket, err)
	} else {
		syscall.Umask(originalUmask)
		log.Infof(s.ctx, "Dapr memory state store listening on %s...\n", socket)
		server := grpc.NewServer()
		proto.RegisterStateStoreServer(server, &srv)
		return server.Serve(listener)
	}
}

func (s *MemoryStore) Init(_ state.Metadata) error {
	for k := range s.data {
		delete(s.data, k)
	}
	return nil
}

func (s *MemoryStore) Features() []state.Feature {
	return []state.Feature{}
}

func (s *MemoryStore) Delete(req *state.DeleteRequest) error {
	delete(s.data, req.Key)
	return nil
}

func (s *MemoryStore) Get(req *state.GetRequest) (*state.GetResponse, error) {
	emptyResponse := &state.GetResponse{
		Data:     nil,
		ETag:     nil,
		Metadata: nil,
	}
	if req == nil {

		return emptyResponse, nil
	}

	log.Debugf(s.ctx, "Get data for %s", req.Key)

	value, ok := s.data[req.Key]
	if !ok {

		return emptyResponse, nil
	}

	metadata := map[string]string{}
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	var etag *string

	return &state.GetResponse{
		Data:     value,
		ETag:     etag,
		Metadata: map[string]string{},
	}, nil
}

func (s *MemoryStore) Set(req *state.SetRequest) error {

	var bytes []byte
	log.Debugf(s.ctx, "Set data for %s", req.Key)

	switch t := req.Value.(type) {
	case string:
		bytes = []byte(t)
	case []byte:
		bytes = t
	default:
		if t == nil {
			return fmt.Errorf("set: request body is nil")
		}
		var err error
		if bytes, err = utils.Marshal(t, json.Marshal); err != nil {
			return err
		}
	}
	log.Debugf(s.ctx, "Stored %d bytes at %s", len(bytes), req.Key)

	s.data[req.Key] = bytes

	return nil
}

func (s *MemoryStore) Ping() error {
	log.Debugf(s.ctx, "PONG!")
	return nil
}

func (s *MemoryStore) BulkDelete(_ []state.DeleteRequest) error {
	log.Debugf(s.ctx, "No bulk delete, sorry!!")
	return nil
}

func (s *MemoryStore) BulkGet(_ []state.GetRequest) (bool, []state.BulkGetResponse, error) {
	log.Debugf(s.ctx, "No bulk get, sorry!!")
	return false, nil, nil
}

func (s *MemoryStore) BulkSet(req []state.SetRequest) error {
	log.Debugf(s.ctx, "Bulk set for %d entities", len(req))
	for _, r := range req {
		err := s.Set(&r)
		if err != nil {
			return err
		}
	}
	return nil
}
