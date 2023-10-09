# Startup of SSCS

SSCS is started as a Daemon. Which keeps running on the system, unless
the network service fails for some reason. 

The basic manual startup process goes like this:

```
make build

sudo ./sscs install

sudo ./sscs start
```

If you are using systemd, the procedure above will create a system unit on 
`/etc/systemd/system/`. You can check if the sscs service is enabled, or installed with:

```
systemctl list-unit-files --type=service  | grep -i sscs
```

##  Other useful commands for debugging 

To check status of service:
```
systemctl status sscs
```

To check all logs of service:
```
sudo journalctl -u sscs -p debug
```

