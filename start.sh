#!/bin/bash
rm -rf deploy
cp -r template deploy

echo "Building server .go file..."
go build -o ./deploy/tcp-server ./tcp-server/*.go 
go build -o ./deploy/http-server ./http-server/*.go

echo "Starting Redis & MySQL Docker..."
docker-compose -f ./docker/docker-compose.yml up -d
sleep 10

cd ./deploy

echo "Starting tcp server..."
nohup ./tcp-server > tcp.nohup.out &
sleep 2

echo "Starting http server..."
nohup ./http-server > http.nohup.out &

echo "Deploy complete!"
echo "Find Pid to kill: sudo lsof -i :8080"