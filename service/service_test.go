package service

import (
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"user/service/api"
	"user/service/mocks"
	"user/service/model"
)

// TESTS
var service *Service
var ctx context.Context

const notExError = "Not expected error: "

type serviceMocks struct {
	RepositoryInterface *mocks.RepositoryInterface
	ProducerInterface   *mocks.ProducerInterface
}

func setupService() (*serviceMocks, *Service) {
	serviceMocks := &serviceMocks{
		RepositoryInterface: new(mocks.RepositoryInterface),
		ProducerInterface:   new(mocks.ProducerInterface),
	}
	service = New(serviceMocks.RepositoryInterface, serviceMocks.ProducerInterface)
	return serviceMocks, service
}

// STATUS ENDPOINT TESTS
func TestServiceGetStatus(t *testing.T) {
	_, testingService := setupService()
	// run test and validate
	reply, err := testingService.GetStatus(ctx, &empty.Empty{})
	if err != nil {
		t.Error(notExError + err.Error())
	}
	assert.Equal(t, api.ServiceStatus_UP, reply.Status)
}

// CREATE USER ENDPOINT TESTS
func TestServiceCreateUserOk(t *testing.T) {
	request := &api.CreateUserRequest{
		Firstname: "firstname",
		Lastname:  "lastname",
		Nickname:  "nickname",
		Email:     "test@email.com",
		Password:  "my_test_password",
		Country:   api.Country_EN,
	}
	createdUserMock := &model.User{
		ID:        "0c10a807-1d58-426a-899d-9ccf8fe57a63",
		Firstname: "firstname",
		Lastname:  "lastname",
		Nickname:  "nickname",
		Email:     "test@email.com",
		Country:   "EN",
		CreatedAt: "creation_date",
		UpdatedAt: "creation_date",
	}
	kafkaMessage := "Created user " + createdUserMock.ID
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("CreateUser", ctx, request).Return(createdUserMock, nil)
	mockServices.ProducerInterface.On("PublishMessage", kafkaMessage).Return(nil)
	// run test and validate
	reply, err := testingService.CreateUser(ctx, request)
	if err != nil {
		t.Error(notExError + err.Error())
	}
	assert.Equal(t, createdUserMock.ID, reply.User.Id)
	assert.Equal(t, createdUserMock.Firstname, reply.User.Firstname)
	assert.Equal(t, createdUserMock.Lastname, reply.User.Lastname)
	assert.Equal(t, createdUserMock.Nickname, reply.User.Nickname)
	assert.Equal(t, createdUserMock.Email, reply.User.Email)
	assert.Equal(t, createdUserMock.Country, reply.User.Country.String())
}

func TestServiceCreateUserInvalidCountryKo(t *testing.T) {
	request := &api.CreateUserRequest{
		Firstname: "firstname",
		Lastname:  "lastname",
		Nickname:  "nickname",
		Email:     "test@email.com",
		Password:  "my_test_password",
		Country:   api.Country_UNKNOWN,
	}
	_, testingService := setupService()
	// run test and validate
	reply, err := testingService.CreateUser(ctx, request)
	if err == nil {
		t.Error(notExError + err.Error())
	}
	assert.Nil(t, reply)
	assertStatusError(t, err, "Received country is not valid", codes.InvalidArgument)
}

func TestServiceCreateUserRepositoryErrorKo(t *testing.T) {
	request := &api.CreateUserRequest{
		Firstname: "firstname",
		Lastname:  "lastname",
		Nickname:  "nickname",
		Email:     "test@email.com",
		Password:  "my_test_password",
		Country:   api.Country_EN,
	}
	error := errors.New("repository error")
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("CreateUser", ctx, request).Return(nil, error)
	// run test and validate
	reply, err := testingService.CreateUser(ctx, request)
	assert.Nil(t, reply)
	if err == nil {
		t.Error(notExError + err.Error())
	}
	assert.Nil(t, reply)
	assertStatusError(t, err, "repository error", codes.Internal)
}

func TestServiceCreateUserProducerErrorKo(t *testing.T) {
	request := &api.CreateUserRequest{
		Firstname: "firstname",
		Lastname:  "lastname",
		Nickname:  "nickname",
		Email:     "test@email.com",
		Password:  "my_test_password",
		Country:   api.Country_EN,
	}
	createdUserMock := &model.User{
		ID:        "0c10a807-1d58-426a-899d-9ccf8fe57a63",
		Firstname: "firstname",
		Lastname:  "lastname",
		Nickname:  "nickname",
		Email:     "test@email.com",
		Country:   "EN",
		CreatedAt: "creation_date",
		UpdatedAt: "creation_date",
	}
	kafkaMessage := "Created user " + createdUserMock.ID
	error := errors.New("producer error")
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("CreateUser", ctx, request).Return(createdUserMock, nil)
	mockServices.ProducerInterface.On("PublishMessage", kafkaMessage).Return(error)
	// run test and validate
	reply, err := testingService.CreateUser(ctx, request)
	if err != nil {
		t.Error(notExError + err.Error())
	}
	assert.Equal(t, createdUserMock.ID, reply.User.Id)
	assert.Equal(t, createdUserMock.Firstname, reply.User.Firstname)
	assert.Equal(t, createdUserMock.Lastname, reply.User.Lastname)
	assert.Equal(t, createdUserMock.Nickname, reply.User.Nickname)
	assert.Equal(t, createdUserMock.Email, reply.User.Email)
	assert.Equal(t, createdUserMock.Country, reply.User.Country.String())
}

// GET ENDPOINT TESTS
func TestGetUsersOk(t *testing.T) {
	country := api.Country_IT
	request := &api.GetUsersRequest{
		FilterCountry: &country,
		Page:          1,
		PageSize:      10,
	}
	decodedUsers := createDecodedUsers()
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("GetUsersPaginated", ctx, request).Return(decodedUsers, nil)
	// run test and validate
	reply, err := testingService.GetUsers(ctx, request)
	if err != nil {
		t.Error(notExError + err.Error())
	}
	assert.NotNil(t, reply)
	assert.Equal(t, int64(1), reply.Page)
	assert.Equal(t, int64(10), reply.PageSize)
	assert.Equal(t, 2, len(reply.Results))
	// Analyze first user
	assert.Equal(t, "1b8b24f8-a56b-4665-88f2-44e144389ce0", reply.Results[0].Id)
	assert.Equal(t, "User 1", reply.Results[0].Firstname)
	assert.Equal(t, "User 1 Lastname", reply.Results[0].Lastname)
	assert.Equal(t, "User 1 Nickname", reply.Results[0].Nickname)
	assert.Equal(t, "user1@test.com", reply.Results[0].Email)
	assert.Equal(t, "IT", reply.Results[0].Country.String())
	assert.Equal(t, "now", reply.Results[0].CreatedAt)
	assert.Equal(t, "now", reply.Results[0].UpdatedAt)
	// Analyze second user
	assert.Equal(t, "3bacc2e9-089a-4c27-b662-d3826b68173b", reply.Results[1].Id)
	assert.Equal(t, "User 2", reply.Results[1].Firstname)
	assert.Equal(t, "User 2 Lastname", reply.Results[1].Lastname)
	assert.Equal(t, "User 2 Nickname", reply.Results[1].Nickname)
	assert.Equal(t, "user2@test.com", reply.Results[1].Email)
	assert.Equal(t, "EN", reply.Results[1].Country.String())
	assert.Equal(t, "now", reply.Results[1].CreatedAt)
	assert.Equal(t, "now", reply.Results[1].UpdatedAt)
}

func TestServiceGetUsersRepositoryErrorKo(t *testing.T) {
	country := api.Country_IT
	request := &api.GetUsersRequest{
		FilterCountry: &country,
		Page:          1,
		PageSize:      10,
	}
	error := errors.New("repository error")
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("GetUsersPaginated", ctx, request).Return(nil, error)
	// run test and validate
	reply, err := testingService.GetUsers(ctx, request)
	assert.Nil(t, reply)
	if err == nil {
		t.Error(notExError + err.Error())
	}
	assert.Nil(t, reply)
	assertStatusError(t, err, "repository error", codes.Internal)
}

// UPDATE ENDPOINT TESTS
func TestUpdateUserOk(t *testing.T) {
	existingUser := createDecodedUsers()[0]
	firstname := "firstname"
	lastname := "lastname"
	nickname := "nickname"
	email := "test@email.com"
	country := api.Country_IT
	request := &api.UpdateUserRequest{
		Id:        "1b8b24f8-a56b-4665-88f2-44e144389ce0",
		Firstname: &firstname,
		Lastname:  &lastname,
		Nickname:  &nickname,
		Email:     &email,
		Country:   &country,
	}
	updatedUser(&existingUser, request)
	kafkaMessage := "Updated user " + existingUser.ID
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("UpdateUser", ctx, request).Return(&existingUser, nil)
	mockServices.ProducerInterface.On("PublishMessage", kafkaMessage).Return(nil)
	// run test and validate
	reply, err := testingService.UpdateUser(ctx, request)
	if err != nil {
		t.Error(notExError + err.Error())
	}
	assert.Equal(t, &empty.Empty{}, reply)
}

func TestServiceUpdateUserInvalidCountryKo(t *testing.T) {
	firstname := "firstname"
	lastname := "lastname"
	nickname := "nickname"
	email := "test@email.com"
	country := api.Country_UNKNOWN
	request := &api.UpdateUserRequest{
		Id:        "1b8b24f8-a56b-4665-88f2-44e144389ce0",
		Firstname: &firstname,
		Lastname:  &lastname,
		Nickname:  &nickname,
		Email:     &email,
		Country:   &country,
	}
	_, testingService := setupService()
	// run test and validate
	reply, err := testingService.UpdateUser(ctx, request)
	if err == nil {
		t.Error(notExError + err.Error())
	}
	assert.Nil(t, reply)
	assertStatusError(t, err, "Received country is not valid", codes.InvalidArgument)
}

func TestServiceUpdateUserRepositoryErrorKo(t *testing.T) {
	firstname := "firstname"
	lastname := "lastname"
	nickname := "nickname"
	email := "test@email.com"
	country := api.Country_IT
	request := &api.UpdateUserRequest{
		Id:        "1b8b24f8-a56b-4665-88f2-44e144389ce0",
		Firstname: &firstname,
		Lastname:  &lastname,
		Nickname:  &nickname,
		Email:     &email,
		Country:   &country,
	}
	error := errors.New("repository error")
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("UpdateUser", ctx, request).Return(nil, error)
	// run test and validate
	reply, err := testingService.UpdateUser(ctx, request)
	assert.Nil(t, reply)
	if err == nil {
		t.Error(notExError + err.Error())
	}
	assert.Nil(t, reply)
	assertStatusError(t, err, "repository error", codes.Internal)
}

func TestServiceUpdateUserProducerErrorKo(t *testing.T) {
	existingUser := createDecodedUsers()[0]
	firstname := "firstname"
	lastname := "lastname"
	nickname := "nickname"
	email := "test@email.com"
	country := api.Country_IT
	request := &api.UpdateUserRequest{
		Id:        "1b8b24f8-a56b-4665-88f2-44e144389ce0",
		Firstname: &firstname,
		Lastname:  &lastname,
		Nickname:  &nickname,
		Email:     &email,
		Country:   &country,
	}
	error := errors.New("producer error")
	updatedUser(&existingUser, request)
	kafkaMessage := "Updated user " + existingUser.ID
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("UpdateUser", ctx, request).Return(&existingUser, nil)
	mockServices.ProducerInterface.On("PublishMessage", kafkaMessage).Return(error)
	// run test and validate
	reply, err := testingService.UpdateUser(ctx, request)
	if err != nil {
		t.Error(notExError + err.Error())
	}
	assert.Equal(t, &empty.Empty{}, reply)
}

// DELETE ENDPOINT TESTS
func TestServiceDeleteUserOk(t *testing.T) {
	request := &api.DeleteUserRequest{
		Id: "0c10a807-1d58-426a-899d-9ccf8fe57a63",
	}
	kafkaMessage := "Deleted user " + request.Id
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("DeleteUser", ctx, request).Return(nil)
	mockServices.ProducerInterface.On("PublishMessage", kafkaMessage).Return(nil)
	// run test and validate
	reply, err := testingService.DeleteUser(ctx, request)
	if err != nil {
		t.Error(notExError + err.Error())
	}
	assert.Equal(t, &empty.Empty{}, reply)
}

func TestServiceDeleteUserNotFoundKo(t *testing.T) {
	request := &api.DeleteUserRequest{
		Id: "AnIdThatDoesNotExists",
	}
	error := errors.New("user AnIdThatDoesNotExists not found")
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("DeleteUser", ctx, request).Return(error)
	// run test and validate
	reply, err := testingService.DeleteUser(ctx, request)
	if err == nil {
		t.Error(notExError + err.Error())
	}
	assert.Nil(t, reply)
	assertStatusError(t, err, "user AnIdThatDoesNotExists not found", codes.NotFound)
}

func TestServiceDeleteUserProducerErrorKo(t *testing.T) {
	request := &api.DeleteUserRequest{
		Id: "0c10a807-1d58-426a-899d-9ccf8fe57a63",
	}
	kafkaMessage := "Deleted user " + request.Id
	error := errors.New("producer error")
	mockServices, testingService := setupService()
	mockServices.RepositoryInterface.On("DeleteUser", ctx, request).Return(nil)
	mockServices.ProducerInterface.On("PublishMessage", kafkaMessage).Return(error)
	// run test and validate
	reply, err := testingService.DeleteUser(ctx, request)
	if err != nil {
		t.Error(notExError + err.Error())
	}
	assert.Equal(t, &empty.Empty{}, reply)
}

// Utility
func assertStatusError(t *testing.T, err error, expectedErrorMessage string, expectedCode codes.Code) {
	statusErr := status.Convert(err)
	// Assert
	assert.Equal(t, expectedCode, statusErr.Code())
	assert.Equal(t, expectedErrorMessage, statusErr.Message())
}

func createDecodedUsers() []model.User {
	var decodedUsers []model.User
	user1 := &model.User{
		ID:        "1b8b24f8-a56b-4665-88f2-44e144389ce0",
		Firstname: "User 1",
		Lastname:  "User 1 Lastname",
		Nickname:  "User 1 Nickname",
		Email:     "user1@test.com",
		Password:  "password",
		Country:   "IT",
		CreatedAt: "now",
		UpdatedAt: "now",
	}
	user2 := &model.User{
		ID:        "3bacc2e9-089a-4c27-b662-d3826b68173b",
		Firstname: "User 2",
		Lastname:  "User 2 Lastname",
		Nickname:  "User 2 Nickname",
		Email:     "user2@test.com",
		Password:  "password",
		Country:   "EN",
		CreatedAt: "now",
		UpdatedAt: "now",
	}
	decodedUsers = append(decodedUsers, *user1)
	decodedUsers = append(decodedUsers, *user2)
	return decodedUsers
}

func updatedUser(userToUpdate *model.User, request *api.UpdateUserRequest) *model.User {
	userToUpdate.Firstname = request.GetFirstname()
	userToUpdate.Lastname = request.GetLastname()
	userToUpdate.Nickname = request.GetNickname()
	userToUpdate.Country = request.Country.String()
	userToUpdate.Email = request.GetEmail()
	userToUpdate.CreatedAt = "created now"
	userToUpdate.UpdatedAt = "updated now"
	return userToUpdate
}
