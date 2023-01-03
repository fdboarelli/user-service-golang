echo "---------INTEGRATION TESTS START---------"
echo "1 - Building service updated image"
cd ../
docker build -t user-service-local .
echo "2 - Starting dockerized environment"
docker-compose up -d
echo "3 - Giving service time to start"
sleep 18
echo "4 - Running go test"
cd client/
go test status_test.go user_crud_flow_test.go integration_test_helper.go
echo "5 - Stopping dockerized environment"
docker-compose stop
echo "---------INTEGRATION TESTS STOP---------"