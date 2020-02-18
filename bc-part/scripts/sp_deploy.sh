#!/usr/bin/env bash

ssh 192.168.77.47 "cd ~/hlf-iot-bc/; git pull; make clean generate up seed"
