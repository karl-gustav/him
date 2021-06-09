package main

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

const (
	masterStorageKey = "him/pick-up-dates"
	gcpProject       = "my-cloud-collection"
)

var loc *time.Location

func init() {
	var err error
	loc, err = time.LoadLocation("Europe/Oslo")
	if err != nil {
		panic(err)
	}
}

func storePickUp(ctx context.Context, data []HIM) error {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return err
	}
	collection := client.Doc(masterStorageKey)
	m := make(map[string]HIM)
	for _, value := range data {
		m[value.GarbageType] = value
	}
	_, err = collection.Set(ctx, m)
	return err
}

func getPickUp(ctx context.Context) ([]HIM, error) {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return nil, err
	}
	containerRef := client.Doc(masterStorageKey)
	document, err := containerRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	var container map[string]HIM
	err = document.DataTo(&container)
	if err != nil {
		return nil, err
	}
	var data []HIM
	for _, value := range container {
		// re-add timezone info because that is lost in firebase
		value.NextDate = value.NextDate.In(loc)
		data = append(data, value)
	}
	return data, nil
}
