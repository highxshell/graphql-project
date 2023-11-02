package database

import (
	"context"
	"gql/graph/model"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var connectionString string = "mongodb://localhost:27017"


type DB struct {
	client *mongo.Client
}

func Connect() *DB{
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil{
		log.Fatal(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil{
		log.Println("Failed to connect to DB!")
		return nil
	}
	log.Println("Connected to DB!")
	return &DB{client: client}
}

func (db *DB) GetJob(id string) *model.JobListing{
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id":_id}
	var jobListing model.JobListing
	err := jobCollection.FindOne(ctx, filter).Decode(&jobListing)
	if err != nil{
		log.Fatal(err)
	}

	return &jobListing
}

func (db *DB) GetJobs() []*model.JobListing{
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	var jobListings []*model.JobListing
	cursor, err := jobCollection.Find(ctx, bson.D{})
	if err != nil{
		log.Fatal(err)
	}
	if err = cursor.All(context.TODO(), &jobListings); err != nil{
		panic(err)
	}

	return jobListings
}

func (db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing{
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	inserted, err := jobCollection.InsertOne(ctx, bson.M{
		"title": 		jobInfo.Title,
		"description": 	jobInfo.Description,
		"url": 			jobInfo.URL,
		"company": 		jobInfo.Company,
	})
	if err != nil{
		log.Fatal(err)
	}

	insertedId := inserted.InsertedID.(primitive.ObjectID).Hex()
	returnJobListing := model.JobListing{ID: insertedId,
		Title: jobInfo.Title,
		Company: jobInfo.Company,
		Description: jobInfo.Description,
		URL: jobInfo.URL,
	}
	return &returnJobListing
}

func (db *DB) UpdateJobListing(jobId string, jobInfo model.UpdateJobListingInput) *model.JobListing{
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	updateJobInfo := bson.M{}

	if jobInfo.Title != nil{
		updateJobInfo["title"] = jobInfo.Title
	}
	if jobInfo.Description != nil{
		updateJobInfo["description"] = jobInfo.Description
	}
	if jobInfo.URL != nil{
		updateJobInfo["url"] = jobInfo.URL
	}

	_id, _ := primitive.ObjectIDFromHex(jobId)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set":updateJobInfo}

	results := jobCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))

	var jobListing model.JobListing
	if err := results.Decode(&jobListing); err != nil{
		log.Fatal(err)
	}
	return &jobListing
}

func (db *DB) DeleteJobListing(jobId string) *model.DeleteJobResponse{
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(jobId)
	filter := bson.M{"_id": _id}
	_, err := jobCollection.DeleteOne(ctx, filter)
	if err != nil{
		log.Fatal(err)
	}

	return &model.DeleteJobResponse{DeletedJobID: jobId}
}
