package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"user/service/api"
)

func TestUserServiceCrudFlow(t *testing.T) {
	client, ctx, cancel, conn := CreateConnWithMetadata(t)
	defer CloseConnection(t, cancel, conn)
	// Create a user
	request := &api.CreateUserRequest{
		Firstname: "firstname",
		Lastname:  "lastname",
		Nickname:  "nickname",
		Email:     "test@email.com",
		Password:  "my_test_password",
		Country:   api.Country_EN,
	}
	response, err := client.CreateUser(ctx, request)
	if err != nil {
		t.Fatalf("GRPC call failed: %v", err)
	}
	createdUser := &api.User{
		Firstname: "firstname",
		Lastname:  "lastname",
		Nickname:  "nickname",
		Email:     "test@email.com",
		Country:   api.Country(api.Country_value["EN"]),
		CreatedAt: "Not to be checked",
		UpdatedAt: "Not to be checked",
	}
	assert.NotNil(t, response)
	assert.Equal(t, createdUser.Firstname, response.User.Firstname)
	assert.Equal(t, createdUser.Lastname, response.User.Lastname)
	assert.Equal(t, createdUser.Nickname, response.User.Nickname)
	assert.Equal(t, createdUser.Email, response.User.Email)
	assert.Equal(t, createdUser.Country, response.User.Country)
	user1Id := response.User.Id
	// Get users
	country := api.Country_EN
	getRequest := &api.GetUsersRequest{
		FilterCountry: &country,
		Page:          0,
		PageSize:      10,
	}
	getResponse, getError := client.GetUsers(ctx, getRequest)
	if getError != nil {
		t.Fatalf("Get GRPC call failed: %v", err)
	}
	assert.NotNil(t, getResponse)
	assert.Equal(t, 1, len(getResponse.Results))
	assert.Equal(t, user1Id, getResponse.Results[0].Id)
	// Update the user
	updatedFirstname := "Ufirstname"
	updatedLastname := "Ulastname"
	updatedNickname := "Unickname"
	updatedEmail := "utest@email.com"
	updatedPassword := "my_new_test_password"
	updatedCountry := api.Country_IT
	updateRequest := &api.UpdateUserRequest{
		Id:        user1Id,
		Firstname: &updatedFirstname,
		Lastname:  &updatedLastname,
		Nickname:  &updatedNickname,
		Email:     &updatedEmail,
		Password:  &updatedPassword,
		Country:   &updatedCountry,
	}
	updateResponse, err := client.UpdateUser(ctx, updateRequest)
	if err != nil {
		t.Fatalf("GRPC call failed: %v", err)
	}
	assert.NotNil(t, updateResponse)
	// Get the updated user with filter on language = IT
	countryIT := api.Country_IT
	getRequestWithItFilterRequest := &api.GetUsersRequest{
		FilterCountry: &countryIT,
		Page:          0,
		PageSize:      10,
	}
	getResponseItFilterResponse, getErrorItFilter := client.GetUsers(ctx, getRequestWithItFilterRequest)
	if getErrorItFilter != nil {
		t.Fatalf("Get with IT filter GRPC call failed: %v", err)
	}
	assert.NotNil(t, getResponseItFilterResponse)
	assert.Equal(t, int64(10), getResponseItFilterResponse.PageSize)
	assert.Equal(t, 1, len(getResponseItFilterResponse.Results))
	assert.Equal(t, user1Id, getResponseItFilterResponse.Results[0].Id)
	assert.Equal(t, updatedFirstname, getResponseItFilterResponse.Results[0].Firstname)
	assert.Equal(t, updatedLastname, getResponseItFilterResponse.Results[0].Lastname)
	assert.Equal(t, updatedEmail, getResponseItFilterResponse.Results[0].Email)
	assert.Equal(t, updatedNickname, getResponseItFilterResponse.Results[0].Nickname)
	assert.Equal(t, countryIT, getResponseItFilterResponse.Results[0].Country)
	// Get empty list with filter on language = EN
	getRequestWithEnFilter := &api.GetUsersRequest{
		FilterCountry: &country,
		Page:          0,
		PageSize:      10,
	}
	getResponseWithFilterEN, getError := client.GetUsers(ctx, getRequestWithEnFilter)
	if getError != nil {
		t.Fatalf("Get GRPC call failed: %v", err)
	}
	assert.NotNil(t, getResponseWithFilterEN)
	assert.Nil(t, getResponseWithFilterEN.Results)
	// Get the updated user with no filter
	getRequestWithNoFilterRequest := &api.GetUsersRequest{
		Page:     0,
		PageSize: 10,
	}
	getResponseNoFilter, getErrorItFilter := client.GetUsers(ctx, getRequestWithNoFilterRequest)
	if getErrorItFilter != nil {
		t.Fatalf("Get with IT filter GRPC call failed: %v", err)
	}
	assert.NotNil(t, getResponseNoFilter)
	assert.Equal(t, int64(10), getResponseNoFilter.PageSize)
	assert.Equal(t, 1, len(getResponseNoFilter.Results))
	assert.Equal(t, user1Id, getResponseNoFilter.Results[0].Id)
	assert.Equal(t, updatedFirstname, getResponseNoFilter.Results[0].Firstname)
	assert.Equal(t, updatedLastname, getResponseNoFilter.Results[0].Lastname)
	assert.Equal(t, updatedEmail, getResponseNoFilter.Results[0].Email)
	assert.Equal(t, updatedNickname, getResponseNoFilter.Results[0].Nickname)
	assert.Equal(t, countryIT, getResponseNoFilter.Results[0].Country)
	// Delete the user
	deleteRequest := &api.DeleteUserRequest{
		Id: user1Id,
	}
	deleteResponse, deleteErr := client.DeleteUser(ctx, deleteRequest)
	if deleteErr != nil {
		t.Fatalf("Delete GRPC call failed: %v", err)
	}
	assert.NotNil(t, deleteResponse)
	// Get users empty after delete
	getResponseAfter, getError := client.GetUsers(ctx, getRequest)
	if getError != nil {
		t.Fatalf("Get GRPC call failed: %v", err)
	}
	assert.NotNil(t, getResponseAfter)
	assert.Nil(t, getResponseAfter.Results)
}
