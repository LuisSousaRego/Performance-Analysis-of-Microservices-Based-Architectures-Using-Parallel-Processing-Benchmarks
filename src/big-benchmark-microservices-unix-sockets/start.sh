rm ./bin/*
rm ./logs/*
go build -o ./bin/core ./core.go
go build -o ./bin/worker ./worker.go


neighbours=10
neighbourhoodSize=10
pingLimit=50000
coreAddr=/tmp/core.sock

echo "neighbours number: $neighbours"
echo "neighbourhood size: $neighbourhoodSize"
echo "ping limit: $pingLimit"

./bin/core -nn $neighbours -ns $neighbourhoodSize -ca $coreAddr &> ./logs/core.log &
CORE_PID=$!

sleep 1

for i in $(seq 1 $((neighbours*neighbourhoodSize)))
do
    ./bin/worker -ca $coreAddr -id $i -pl $pingLimit &> ./logs/worker$i.log &
done

wait $CORE_PID

echo "core log:"
cat ./logs/core.log

killall worker