# Example

<div align="center">
    <img height="100px" src="doc/seo.do.png"><br>
    <strong>example</strong>
</div>

## Why

We want to test kafka-do with an end-to-end application.

Used kafka-do version: **`0.1.5`**

## How to use it

This an cli application.  
Usage:  

```
                _
__ ___ __  _ __| |
\ \ / '  \| '_ \ |
/_\_\_|_|_| .__/_|
          |_|
            v0.1.0

Usage: example [options]
Options:
	-w,  --worker-count, *[Default: 100]
	-k,  --kafka         *[Usage: brokers<>inTopic<>outToic. Example: localhost:9094,localhost:9095<>in<>out]	
	-n,  --no-banner
	-v,  --verbose
	-h,  --help
```

Example usage:  
```
example --worker-count 3 \
        --kafka 'kafka:9092<>requests<>responses' \
        --no-banner \
        --verbose
```

## Preview

<div align="center">
    <img src="doc/preview.gif">
</div>

## For development  

Using "dev" you can simulate e2e application usage. It uses tools to send random messages and writing consumed messages to the device.

Run kafka firstly,
```sh
make run-kafka
```

Build images,
```sh
make build-dev
```

Run dev,
```sh
make run-dev
```

To clean dev containers run below commands,  
```
make clean-kafka
```
```
make clean-dev
```

To clean everything,
```
make clean-all
```

## For production  

```
make run-prod
```
