go get -t -d -v ./...
go build -o webapp *.go
sudo rm /var/log/mysql/mysql-slow.log
sudo service mysql restart
sudo service nginx restart

./webapp
