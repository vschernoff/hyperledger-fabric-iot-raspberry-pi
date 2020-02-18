#!/usr/bin/env bash

ssh ubuntu@192.168.77.36 "cd ~/hlf-iot-bc/; git pull; make clean generate up seed"
