# A prometheus exporter for Raspberry pi

## Usage

```sh
go build -o app .
./app --host <host> --port <port> --path.sysfs=/sys
```

Then you can access `/metrics`.

(TBH, it also compatible to anything which has the path "/sys/class/thermal/thermal_zone\*/temp").
