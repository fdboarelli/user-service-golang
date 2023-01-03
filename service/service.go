package service

// This service implements the user_service grpc server business logic
//go:generate mockery --all --output $PWD/mocks
import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"user/service/api"
	"user/service/model"
)

// RepositoryInterface defines the operations exposed by the data layer regarding users
type RepositoryInterface interface {
	CreateUser(ctx context.Context, request *api.CreateUserRequest) (*model.User, error)
	GetUsersPaginated(ctx context.Context, request *api.GetUsersRequest) ([]model.User, error)
	UpdateUser(ctx context.Context, request *api.UpdateUserRequest) (*model.User, error)
	DeleteUser(ctx context.Context, request *api.DeleteUserRequest) error
}

type ProducerInterface interface {
	PublishMessage(message string) error
}

// Service defines the Service composition
type Service struct {
	RepositoryInterface RepositoryInterface
	ProducerInterface   ProducerInterface
}

// New allows to create a new instance of the Service
func New(repository RepositoryInterface, producer ProducerInterface) *Service {
	return &Service{repository, producer}
}

// GetStatus implements the Service's status endpoint, useful for probes and monitoring
func (s *Service) GetStatus(ctx context.Context, e *empty.Empty) (*api.StatusReply, error) {
	return &api.StatusReply{
		Status:  api.ServiceStatus_UP,
		Message: "Account service up and running",
	}, nil
}

// CreateUser allows creation of a new user with given parameters
func (s *Service) CreateUser(ctx context.Context, request *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	log.Info("Starting create user for ", request.Firstname)
	if request.Country.Number() == 0 {
		// Country value is not valid
		log.Info("Received country is not valid ", request.Country)
		return nil, status.Error(codes.InvalidArgument, "Received country is not valid")
	}
	user, err := s.RepositoryInterface.CreateUser(ctx, request)
	if err != nil {
		log.Error("Failed to create user ", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Send event to topic to notify other services
	err = s.ProducerInterface.PublishMessage("Created user " + user.ID)
	if err != nil {
		log.Error("Failed to send create message user")
	}
	log.Info("User created with id ", user.ID)
	return &api.CreateUserResponse{
		User: &api.User{
			Id:        user.ID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Nickname:  user.Nickname,
			Email:     user.Email,
			Country:   api.Country(api.Country_value[user.Country]),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// GetUsers returns a paginated list of users considering given filters (if any, only country filed is supported for now) and page, pageSize parameters
func (s *Service) GetUsers(ctx context.Context, request *api.GetUsersRequest) (*api.GetUserResponse, error) {
	log.Info("Starting get paginated users with parameters ", request.Page, request.PageSize, request.FilterCountry)
	decodedUsers, err := s.RepositoryInterface.GetUsersPaginated(ctx, request)
	if err != nil {
		log.Error("Failed to retrieve paginated users ", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	var grpcUsers []*api.User
	for _, decodedUser := range decodedUsers {
		currentUser := &api.User{
			Id:        decodedUser.ID,
			Firstname: decodedUser.Firstname,
			Lastname:  decodedUser.Lastname,
			Nickname:  decodedUser.Nickname,
			Email:     decodedUser.Email,
			Country:   api.Country(api.Country_value[decodedUser.Country]),
			CreatedAt: decodedUser.CreatedAt,
			UpdatedAt: decodedUser.UpdatedAt,
		}
		grpcUsers = append(grpcUsers, currentUser)
	}
	log.Info("Completed get paginated users")
	return &api.GetUserResponse{
		Results:    grpcUsers,
		Page:       request.Page,
		PageSize:   request.PageSize,
		TotalCount: int64(len(decodedUsers)),
	}, nil
}

// UpdateUser returns and empty body if operation is successful, error otherwise
func (s *Service) UpdateUser(ctx context.Context, request *api.UpdateUserRequest) (*empty.Empty, error) {
	log.Info("Starting update func for user ", request.Id)
	if request.Country.Number() == 0 {
		// Country value is not valid
		log.Info("Received country is not valid ", request.Country)
		return nil, status.Error(codes.InvalidArgument, "Received country is not valid")
	}
	existingUser, err := s.RepositoryInterface.UpdateUser(ctx, request)
	if err != nil {
		log.Error("Failed to update user ", request.Id)
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Send event to topic to notify other services
	err = s.ProducerInterface.PublishMessage("Updated user " + existingUser.ID)
	if err != nil {
		log.Error("Failed to send update message user")
	}
	log.Info("Updated user ", existingUser.ID)
	return &empty.Empty{}, nil
}

// DeleteUser returns and empty body if operation is successful, error otherwise
func (s *Service) DeleteUser(ctx context.Context, request *api.DeleteUserRequest) (*empty.Empty, error) {
	log.Info("Starting deleted func for user", request.Id)
	err := s.RepositoryInterface.DeleteUser(ctx, request)
	if err != nil {
		log.Error("Failed to delete user ", request.Id)
		return nil, status.Error(codes.NotFound, "user "+request.Id+" not found")
	}
	// Send event to topic to notify other services
	err = s.ProducerInterface.PublishMessage("Deleted user " + request.Id)
	if err != nil {
		log.Error("Failed to send create message user")
	}
	log.Info("Deleted user ", request.Id)
	return &empty.Empty{}, nil
}
