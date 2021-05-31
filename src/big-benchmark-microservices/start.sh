echo "building services"
go build -o ./bin/core ./core.go && echo "core built"
go build -o ./bin/worker ./worker.go && echo "worker built"

echo "starting core"
./bin/core &> ./logs/core.log &

sleep 1

echo "starting workers"
for i in {1..2}
do
    ./bin/worker &> ./logs/worker$i.log &
done