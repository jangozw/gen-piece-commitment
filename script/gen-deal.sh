#!/usr/bin/env bash
deal_dir="/Volumes/new-apfs/devnet/deals-data"
src_dir="${deal_dir}/txt"
car_dir="${deal_dir}/car"
miner="t01000"
for n in {0..100}
do
  car_name="${n}.car"
  car="${car_dir}/${car_name}"

  echo $car_name

  lotus client generate-car "${src_dir}/${n}.txt" "$car"
  res=$(lotus client import --car "$car")
  root=$(echo $res| awk '{print $4}')
  echo "root: $root"

  res=$(lotus client commP $car)
  pieceCid=`echo $res | awk '{print $2}'`
  echo "pieceCid: $pieceCid"

  proposoalCid=$(lotus client deal --fast-retrieval=true --manual-stateless-deal=true   --manual-piece-cid="${pieceCid}"  --manual-piece-size=2032 --verified-deal=false  "${root}" "$miner" 0 518400)
  echo "proposoalCid: $proposoalCid"

  echo "$proposoalCid $car_name $miner" >> "${deal_dir}/deals_${miner}.txt"
done
