package repository

// This repository implements the logic to interact with the data layered that has been implemented using MongoDB

import (
	"context"
	"errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"user/service/api"
	"user/service/model"
)

const (
	RFC3339 = "2006-01-02T15:04:05Z07:00"
)

type UtilityInterface interface {
	EncodeInput(input string) string
}

func New(client *mongo.Client, utility UtilityInterface) *Repository {
	return &Repository{client: client, UtilityInterface: utility}
}

type Repository struct {
	client           *mongo.Client
	UtilityInterface UtilityInterface
}

// CreateUser returns the created User
func (repository *Repository) CreateUser(ctx context.Context, request *api.CreateUserRequest) (*model.User, error) {
	log.Debug("Creating new user entity")
	currentTime := time.Now()
	userId := uuid.New().String()
	log.Debug("New user id is", userId)
	creationDate := currentTime.Format(RFC3339)
	hashedPassword := repository.UtilityInterface.EncodeInput(request.Password)
	user := model.User{
		ID:        userId,
		Firstname: request.Firstname,
		Lastname:  request.Lastname,
		Nickname:  request.Nickname,
		Password:  hashedPassword,
		Email:     request.Email,
		Country:   request.Country.String(),
		CreatedAt: creationDate,
		UpdatedAt: creationDate,
	}
	usersCollection := repository.GetConnection()
	_, error := usersCollection.InsertOne(context.TODO(), user)
	if error != nil {
		log.Error("Error while creating the user", error)
		return nil, error
	}
	log.Debug("Generated user with id ", userId)
	return &user, nil
}

// GetUsersPaginated returns a list of users according to given search filters
func (repository *Repository) GetUsersPaginated(ctx context.Context, request *api.GetUsersRequest) ([]model.User, error) {
	log.Debug("Starting paginated retrieval of users")
	usersCollection := repository.GetConnection()
	var decodedUsers []model.User
	var filters []string
	filter := bson.D{}
	// If specific language are passed, use them as filter
	if request.FilterCountry != nil {
		filters = append(filters, request.FilterCountry.String())
		filter = bson.D{{
			Key:   "country",
			Value: bson.D{{Key: "$in", Value: filters}},
		}}
	}
	// Paging options
	pageOptions := options.Find()
	pageOptions.SetSkip(request.Page)      //0-i
	pageOptions.SetLimit(request.PageSize) // number of records to return
	users, err := usersCollection.Find(context.TODO(), filter, pageOptions)
	defer func(users *mongo.Cursor, ctx context.Context) {
		err := users.Close(ctx)
		if err != nil {
			log.Error("Error in close db stream func ", err)
		}
	}(users, ctx)
	// var results []bson.D
	if err = users.All(ctx, &decodedUsers); err != nil {
		log.Error("Error while unmarshalling users data ", err)
		return nil, err
	}
	log.Debug("Retrieved paged users successfully")
	return decodedUsers, nil
}

// UpdateUser returns the updated User
func (repository *Repository) UpdateUser(ctx context.Context, request *api.UpdateUserRequest) (*model.User, error) {
	log.Debug("Starting update user for user ", request.Id)
	usersCollection := repository.GetConnection()
	var existingUser *model.User
	filter := bson.D{{"id", request.Id}}
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err != nil {
		log.Error("Error while getting the use with id ", request.Id)
		log.Error(err)
		return nil, err
	}
	if request.Firstname != nil {
		log.Debug("Firstname is update from ", existingUser.Firstname, " to ", request.Firstname)
		existingUser.Firstname = request.GetFirstname()
	}
	if request.Lastname != nil {
		log.Debug("Lastname is update from ", existingUser.Lastname, " to ", request.Lastname)
		existingUser.Lastname = request.GetLastname()
	}
	if request.Nickname != nil {
		log.Debug("Nickname is update from ", existingUser.Nickname, " to ", request.Nickname)
		existingUser.Nickname = request.GetNickname()
	}
	if request.Email != nil {
		log.Debug("Email is update from ", existingUser.Email, " to ", request.Email)
		existingUser.Email = request.GetEmail()
	}
	if request.Password != nil {
		hashedPassword := repository.UtilityInterface.EncodeInput(request.GetPassword())
		log.Debug("Password is updated")
		existingUser.Password = hashedPassword
	}
	if request.Country != nil {
		log.Debug("Country is update from ", existingUser.Country, " to ", request.Country)
		existingUser.Country = request.GetCountry().String()
	}
	currentTime := time.Now()
	updatedAtDate := currentTime.Format(RFC3339)
	existingUser.UpdatedAt = updatedAtDate
	updateFilter := bson.D{{"$set", existingUser}}
	_, err = usersCollection.UpdateOne(ctx, bson.M{}, updateFilter)
	if err != nil {
		log.Error("Error while getting the use with id ", request.Id)
		log.Error(err)
		return nil, err
	}
	log.Debug("Updated users ", existingUser.ID)
	return existingUser, nil
}

// DeleteUser delete a user, returns an error if operation fails
func (repository *Repository) DeleteUser(ctx context.Context, request *api.DeleteUserRequest) error {
	log.Debug("Starting deletion func for user ", request.Id)
	usersCollection := repository.GetConnection()
	filter := bson.M{"id": request.Id}
	deleteResult, err := usersCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Error("Cannot delete user ", request.Id)
		log.Error(err)
		return err
	}
	if deleteResult.DeletedCount == 0 {
		log.Error("User ", request.Id, " is not present in the database")
		return errors.New("user is not present in the database")
	}
	log.Debug("Correctly deleted user ", request.Id)
	return nil
}

// GetConnection is used to establish the connection to the users collection in the service's database
func (repository *Repository) GetConnection() *mongo.Collection {
	return repository.client.Database("users_collection").Collection("users")
}
