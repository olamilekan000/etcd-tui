package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/olamilekan000/etcd-tui/internal/config"
	"github.com/olamilekan000/etcd-tui/internal/utils"
)

type Repository interface {
	Connect() tea.Msg
	FetchKeys(startKey string, limit int) tea.Cmd
	FetchAllKeys() tea.Cmd
	FetchTotalCount() tea.Cmd
	FetchValue(key string) tea.Cmd
	SetClient(client *clientv3.Client)
	Close() error
}

type repository struct {
	client *clientv3.Client
}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) SetClient(client *clientv3.Client) {
	r.client = client
}

func (r *repository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *repository) Connect() tea.Msg {
	endpoints := config.GetEndpoints()
	if endpoints == "" {
		return ConnectionMsg{Success: false, Err: fmt.Errorf("ETCDCTL_ENDPOINTS not set, check config file or environment variables")}
	}

	caCertPath := config.GetCACert()
	keyPath := config.GetKey()
	certPath := config.GetCert()

	endpointsList := strings.Split(endpoints, ",")

	var tlsConfig *tls.Config

	if caCertPath != "" && keyPath != "" && certPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return ConnectionMsg{Success: false, Err: fmt.Errorf("failed to read CA cert: %w", err)}
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return ConnectionMsg{Success: false, Err: fmt.Errorf("failed to parse CA cert")}
		}

		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return ConnectionMsg{Success: false, Err: fmt.Errorf("failed to load client cert: %w", err)}
		}

		tlsConfig = &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{cert},
		}
	}

	username := config.GetUsername()
	password := config.GetPassword()

	clientConfig := clientv3.Config{
		Endpoints:   endpointsList,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	}

	if username != "" && password != "" {
		clientConfig.Username = username
		clientConfig.Password = password
	}

	client, err := clientv3.New(clientConfig)
	if err != nil {
		return ConnectionMsg{Success: false, Err: fmt.Errorf("failed to create etcd client: %w", err)}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if len(endpointsList) == 0 {
		client.Close()
		return ConnectionMsg{Success: false, Err: fmt.Errorf("no endpoints provided")}
	}

	_, err = client.Status(ctx, endpointsList[0])
	if err != nil {
		client.Close()
		return ConnectionMsg{Success: false, Err: fmt.Errorf("failed to connect to etcd: %w", err)}
	}

	r.client = client

	return ConnectionMsg{Client: client, Success: true}
}

func (r *repository) FetchKeys(startKey string, limit int) tea.Cmd {
	return func() tea.Msg {
		if r.client == nil {
			return KeysMsg{Err: fmt.Errorf("etcd client not initialized")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if limit <= 0 {
			limit = 100
		}

		var opts []clientv3.OpOption
		var queryKey string

		if startKey == "" {
			queryKey = ""
			opts = []clientv3.OpOption{
				clientv3.WithPrefix(),
				clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
				clientv3.WithLimit(int64(limit)),
			}
		} else {
			queryKey = startKey + "\x00"
			opts = []clientv3.OpOption{
				clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
				clientv3.WithLimit(int64(limit)),
				clientv3.WithFromKey(),
			}
		}

		resp, err := r.client.Get(ctx, queryKey, opts...)
		if err != nil {
			return KeysMsg{Err: err}
		}

		kvPairs := make([]KeyValue, 0, len(resp.Kvs))
		skipFirst := false
		if startKey != "" && len(resp.Kvs) > 0 {
			firstKey := utils.SanitizeForTUI(string(resp.Kvs[0].Key))
			if firstKey == startKey {
				skipFirst = true
			}
		}

		for i, kv := range resp.Kvs {
			if skipFirst && i == 0 {
				continue
			}

			keyStr := utils.SanitizeForTUI(string(kv.Key))
			valueStr := utils.SanitizeForTUI(string(kv.Value))
			valueStr = strings.TrimSpace(valueStr)

			var preview string
			if len(valueStr) == 0 {
				preview = "no value"
			} else {
				preview = utils.NormalizeForDisplay(valueStr, 50)
			}

			kvPairs = append(kvPairs, KeyValue{
				Key:          keyStr,
				Value:        valueStr,
				ValuePreview: preview,
			})
		}

		hasMore := len(resp.Kvs) >= limit

		return KeysMsg{
			Keys:    kvPairs,
			HasMore: hasMore,
		}
	}
}

func (r *repository) FetchAllKeys() tea.Cmd {
	return func() tea.Msg {
		if r.client == nil {
			return KeysMsg{Err: fmt.Errorf("etcd client not initialized")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := []clientv3.OpOption{
			clientv3.WithPrefix(),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		}

		resp, err := r.client.Get(ctx, "", opts...)
		if err != nil {
			return KeysMsg{Err: err}
		}

		kvPairs := make([]KeyValue, 0, len(resp.Kvs))

		for _, kv := range resp.Kvs {
			keyStr := utils.SanitizeForTUI(string(kv.Key))
			valueStr := utils.SanitizeForTUI(string(kv.Value))
			valueStr = strings.TrimSpace(valueStr)

			var preview string
			if len(valueStr) == 0 {
				preview = "no value"
			} else {
				preview = utils.NormalizeForDisplay(valueStr, 50)
			}

			kvPairs = append(kvPairs, KeyValue{
				Key:          keyStr,
				Value:        valueStr,
				ValuePreview: preview,
			})
		}

		return KeysMsg{
			Keys:    kvPairs,
			HasMore: false,
		}
	}
}

func (r *repository) FetchTotalCount() tea.Cmd {
	return func() tea.Msg {
		if r.client == nil {
			return CountMsg{Count: -1, Err: fmt.Errorf("etcd client not initialized")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := r.client.Get(ctx, "", clientv3.WithPrefix(), clientv3.WithCountOnly())
		if err != nil {
			return CountMsg{Count: -1, Err: err}
		}

		return CountMsg{Count: int(resp.Count)}
	}
}

func (r *repository) FetchValue(key string) tea.Cmd {
	return func() tea.Msg {
		if r.client == nil {
			return ValueMsg{Key: key, Err: fmt.Errorf("etcd client not initialized")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := r.client.Get(ctx, key)
		if err != nil {
			return ValueMsg{Key: key, Err: err}
		}

		value := ""
		if len(resp.Kvs) > 0 {
			value = utils.SanitizeForTUI(string(resp.Kvs[0].Value))
			value = strings.TrimSpace(value)
		}

		return ValueMsg{Key: key, Value: value}
	}
}
