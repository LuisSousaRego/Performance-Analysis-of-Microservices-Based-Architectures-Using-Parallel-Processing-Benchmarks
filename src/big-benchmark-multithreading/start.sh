go build -o ./bin/main ./main.go && echo "main built"

echo "starting main"
./bin/main &> ./logs/main.log &
