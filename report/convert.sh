chmod +x pandoc.sh

docker run --rm --volume "`pwd`:/data" --entrypoint "/data/pandoc.sh" pandoc/latex:2.14
