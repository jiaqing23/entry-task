#!/bin/bash
rm -rf deploy
cp -r template deploy

sudo service goweb stop

echo "Building server .go file..."
go build -o ./deploy/tcp-server ./tcp-server/*.go 
go build -o ./deploy/http-server ./http-server/*.go
cp ./http-server/.env ./deploy

echo "Starting tcp server..."
nohup ./deploy/tcp-server > tcp.nohup.out &
sleep 2

echo "Starting http server..."
sudo service goweb start

echo "Deploy complete!"
echo "Find pid to kill: sudo lsof -i :8080"
echo ""

echo "------http server status------"
sudo service goweb status
echo "------------------------------"
echo "Connect at: http://10.143.139.181/"


#sudo service goweb stop
#journalctl -u goweb --no-pager --since "2021-08-08 14:15:58"

#sudo service goweb stop
#go build -o ./deploy/http-server ./http-server/*.go
#sudo service goweb start
