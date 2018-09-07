# hunter-agent

```
 __ __  __ __  ____   ______    ___  ____          ____   ____    ___  ____   ______
 |  |  ||  |  ||    \ |      |  /  _]|    \        /    | /    |  /  _]|    \ |      |
 |  |  ||  |  ||  _  ||      | /  [_ |  D  )_____ |  o  ||   __| /  [_ |  _  ||      |
 |  _  ||  |  ||  |  ||_|  |_||    _]|    /|     ||     ||  |  ||    _]|  |  ||_|  |_|
 |  |  ||  :  ||  |  |  |  |  |   [_ |    \|_____||  _  ||  |_ ||   [_ |  |  |  |  |
 |  |  ||     ||  |  |  |  |  |     ||  .  \      |  |  ||     ||     ||  |  |  |  |
 |__|__| \__,_||__|__|  |__|  |_____||__|\_|      |__|__||___,_||_____||__|__|  |__|
```

This is an agent for hunter system as a proxy.

## Design

- The location of hunter agent and the whole architecture

![whole](https://raw.githubusercontent.com/moooofly/hunter-agent/master/docs/whole.png)

- The callchain of business logic

![callchain](https://raw.githubusercontent.com/moooofly/hunter-agent/master/docs/callchain.png)

- The protocol between different components

![protocol](https://raw.githubusercontent.com/moooofly/hunter-agent/master/docs/protocol.png)


## Config

You can put `agent.json.template` into `/etc/hunter/` with name `agent.json`.


## Signal Usage

- SIGHUP: Reload config file
- SIGUSR1: Dump stacks of all goroutines
- SIGPIPE: Ignore
- SIGINT or SIGTERM: Your custom `cleanup` is called, then the process is terminated.
- SIGQUIT: Cause an exit without cleanup, with a goroutine dump preceding exit.
