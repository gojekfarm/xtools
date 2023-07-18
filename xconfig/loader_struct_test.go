package xconfig

import (
	"errors"
	"testing"
)

type App struct {
	Name    string  `config:"NAME"`
	Cluster Cluster `config:",prefix=CLUSTER_"`
}

type Cluster struct {
	Name    string  `config:"NAME"`
	Master  Server  `config:",prefix=MASTER_"`
	Replica *Server `config:",prefix=REPLICA_"`
}

type Server struct {
	Host string `config:"HOST"`
	Port int    `config:"PORT"`
}

type User struct {
	Name  string `json:"NAME"`
	Email string `json:"EMAIL"`
}

func TestLoadWith_Structs(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{
			name:  "nested struct: using prefix",
			input: &App{},
			want: &App{
				Name: "app1",
				Cluster: Cluster{
					Name: "cluster1",
					Master: Server{
						Host: "master1",
						Port: 1,
					},
					Replica: &Server{
						Host: "replica1",
						Port: 2,
					},
				},
			},
			loader: MapLoader{
				"NAME":                 "app1",
				"CLUSTER_NAME":         "cluster1",
				"CLUSTER_MASTER_HOST":  "master1",
				"CLUSTER_MASTER_PORT":  "1",
				"CLUSTER_REPLICA_HOST": "replica1",
				"CLUSTER_REPLICA_PORT": "2",
			},
		},
		{
			name: "nested struct: without prefix",
			input: &struct {
				Name   string `config:"NAME"`
				Server Server
			}{},
			want: &struct {
				Name   string `config:"NAME"`
				Server Server
			}{
				Name: "app1",
				Server: Server{
					Host: "master1",
					Port: 1,
				},
			},
			loader: MapLoader{
				"NAME": "app1",
				"HOST": "master1",
				"PORT": "1",
			},
		},
		{
			name: "non-struct field with prefix",
			input: &struct {
				Name string `config:",prefix=CLUSTER"`
			}{},
			err:    errors.New("prefix is only valid on struct types"),
			loader: MapLoader{},
		},
	}

	runTestcases(t, testcases)
}

func TestLoadWith_JSON(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{
			name: "json object as string",
			input: &struct {
				Owner User `config:"OWNER"`
			}{},
			want: &struct {
				Owner User `config:"OWNER"`
			}{
				Owner: User{
					Name:  "user1",
					Email: "owner@example.com",
				},
			},
			loader: MapLoader{
				"OWNER": `{"NAME":"user1","EMAIL":"owner@example.com"}`,
			},
		},
		{
			name: "json array as string",
			input: &struct {
				Admins []User `config:"ADMINS"`
			}{},
			want: &struct {
				Admins []User `config:"ADMINS"`
			}{
				Admins: []User{
					{Name: "user1", Email: "owner1@example.com"},
					{Name: "user2", Email: "owner2@example.com"},
				},
			},
			loader: MapLoader{
				"ADMINS": `[{"NAME":"user1","EMAIL":"owner1@example.com"},{"NAME":"user2","EMAIL":"owner2@example.com"}]`,
			},
		},
	}

	runTestcases(t, testcases)
}
