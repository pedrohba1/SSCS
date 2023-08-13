# Startup of SCSS

SCSS is started as a Daemon. Which keeps running on the system, unless
the network service fails for some reason. 

The basic manual startup process goes like this:

```
make build

sudo ./scss install

sudo ./scss start
```

If you are using systemd, the procedure above will create a system unit on 
`/etc/systemd/system/`. You can check if the scss service is enabled, or installed with:

```
systemctl list-unit-files --type=service  | grep -i scss
```

##  Other useful commands for debugging 

To check status of service:
```
systemctl status scss
```

To check all logs of service:
```
sudo journalctl -u scss -p debug
```

