sudo service goweb stop
sudo service gotcp stop

sudo lsof -i :3000
sudo lsof -i :8080
echo "Stopped."
