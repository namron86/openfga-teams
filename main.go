package main

import (
	"context"
	"encoding/json"
	"fmt"
	fga "github.com/openfga/go-sdk"
	fgac "github.com/openfga/go-sdk/client"
	"log/slog"
	"os"
)

var l *slog.Logger
var ctx context.Context
var client fgac.SdkClient

func main() {
	l = slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx = context.Background()
	var err error

	client, err = fgac.NewSdkClient(&fgac.ClientConfiguration{
		ApiUrl: "http://localhost:8080",
	})

	authModelId, err := loadModel()
	if err != nil {
		panic(err)
	}

	options := fgac.ClientWriteOptions{
		AuthorizationModelId: fga.PtrString(authModelId),
	}

	writes := fgac.ClientWriteRequest{
		Writes: []fgac.ClientTupleKey{
			{
				User:     "user:Matthew",
				Relation: "owner",
				Object:   "team:A",
			},
			{
				User:     "team:A",
				Relation: "parent",
				Object:   "team:B",
			},
			{
				User:     "team:B",
				Relation: "parent",
				Object:   "team:C",
			},
			{
				User:     "team:C",
				Relation: "parent",
				Object:   "team:D",
			},
		},
	}

	_, err = client.Write(ctx).Body(writes).Options(options).Execute()
	if err != nil {
		panic(err)
	}

	check := fgac.ClientCheckRequest{
		User:     "user:Matthew",
		Relation: "can_read_teams_opportunities",
		Object:   "team:D",
	}

	checkOpts := fgac.ClientCheckOptions{
		AuthorizationModelId: fga.PtrString(authModelId),
	}

	checkData, err := client.Check(ctx).Body(check).Options(checkOpts).Execute()
	if err != nil {
		panic(err)
	}

	l.Info("Can Matthew read Team D's opportunities?", "allowed", checkData.GetAllowed())

	listOpts := fgac.ClientListObjectsOptions{
		AuthorizationModelId: fga.PtrString(authModelId),
	}

	list := fgac.ClientListObjectsRequest{
		User:     "user:Matthew",
		Relation: "can_read_teams_opportunities",
		Type:     "team",
	}

	listData, err := client.ListObjects(ctx).Body(list).Options(listOpts).Execute()
	if err != nil {
		panic(err)
	}

	l.Info("What Teams can Matthew read opportunities from?", "objects", listData.Objects)
}

func loadModel() (string, error) {
	storeResp, err := client.CreateStore(ctx).Body(fgac.ClientCreateStoreRequest{Name: "FGA Demo"}).Execute()
	if err != nil {
		panic(err)
	}

	err = client.SetStoreId(storeResp.Id)
	if err != nil {
		panic(err)
	}

	f := "./model.json"
	s, err := os.ReadFile(f)
	if err != nil {
		panic(fmt.Sprintf("failed to read authorization model file %q: %v", f, err))
	}

	var body fga.WriteAuthorizationModelRequest
	if err = json.Unmarshal(s, &body); err != nil {
		panic(fmt.Sprintf("failed to unmarshal authorization model file %q: %v", f, err))
	}

	modelResp, err := client.WriteAuthorizationModel(ctx).Body(body).Execute()
	if err != nil {
		panic(err)
	}

	return modelResp.AuthorizationModelId, nil
}
