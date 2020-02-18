# raspberry-boot-script

1) Execute command
```
make build
```
2) Add to ~/.bashrc $GOPATH and $GOROOT

For example:
```
export GOPATH=$HOME/Examples
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```
3) Add to /etc/rc.local executing command with path to your binary file

For example:
```
cd /home/pi/Examples/src/hlf-iot-bc-full/raspberry-part/raspberry-boot-script/build
sudo -E ./main
```

before:
```
exit 0
```