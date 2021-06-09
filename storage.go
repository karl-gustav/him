package main

import (
	"context"

	"cloud.google.com/go/firestore"
)

const (
	masterStorageKey = "him/pick-up-dates"
	gcpProject       = "my-cloud-collection"
)

func StorePickUp(ctx context.Context, data []HIM) error {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return err
	}
	collection := client.Doc(masterStorageKey)

	_, err = collection.Set(ctx, data)
	return err
}

func GetPickUp(ctx context.Context) ([]HIM, error) {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return nil, err
	}
	containerRef := client.Doc(masterStorageKey)
	document, err := containerRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	var container []HIM
	err = document.DataTo(&container)
	if err != nil {
		return nil, err
	}
	return container, nil
}
