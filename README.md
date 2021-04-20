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

Tests can be run in the emulator environment and the testnet environment. You first have to set an environment variable to set the environment you want to test on.

On Windows (run cmd as administrator):
```
setx veolettestenv "emulator" / "testnet" /M
```