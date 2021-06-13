go build -o ./bin/core ./core.go
go build -o ./bin/worker ./worker.go

./bin/core &> ./logs/core.log &
CORE_PID=$!

sleep 1

for i in {1..100}
do
    ./bin/worker &> ./logs/worker$i.log &
done

wait $CORE_PID

cat ./logs/core.log