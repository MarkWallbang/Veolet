# Veolet


## Flow emulator

Command to start emulator:

```
$ flow emulator start -v
```

Command to deploy project on emulator:

```
flow project deploy --network=emulator
```

## Start script from CLI

In order to run a script from the CLI, use the following command:

```
flow scripts execute --code=cadence\scripts\Filename.cdc
```

### Script with arguments
```
flow scripts execute --code=cadence\scripts\Tester.cdc --args="[{\"type\":\"String\", \"value\": \"Test\"}]"
```

## Run tests

Tests can be run from inside the cadence/lib/go/test folder. Simply navigate to the folder and run the following command:

```
go test -v
```