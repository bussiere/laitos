# Web service: recurring commands

## Introduction
Hosted by laitos [web server](https://github.com/HouzuoGuo/laitos/wiki/Daemon:-web-server), the service hosts channels
of pre-configured toolbox commands that are run at regular interval, and let user retrieve the command results in JSON
array per channel.

While the service is online, user may add more toolbox commands and put text messages directly into command results via
an HTML form served by this service on the same HTTP endpoint. These transient commands are not memorised and will be
lost upon program restart.

An example use case of the service may be to build a utility web application that displays the latest system resource
usage for monitoring, or the latest list of mails in inbox.

## Configuration
- Under JSON key `HTTPHandlers`, write a string property called `RecurringCommandsEndpoint`, value being the URL
  location that will serve the configuration form and retrieve command results (both under one endpoint). Keep the
  location a secret to yourself and make it difficult to guess.
- Under JSON key `RecurringCommandsEndpointConfig`, create an inner object `RecurringCommands`, in which keys are
  channel names (keep them difficult to guess) and each value is an object with the following mandatory properties: 
<table>
<tr>
    <th>Property</th>
    <th>Type</th>
    <th>Meaning</th>
</tr>
<tr>
    <td>IntervalSec</td>
    <td>integer</td>
    <td>The interval (seconds) at which pre-configured and transient commands toolbox commands should be executed.</td>
</tr>
<tr>
    <td>MaxResults</td>
    <td>integer</td>
    <td>The number of command results to keep. Older results are discarded.</td>
</tr>
<tr>
    <td>PreConfiguredCommands</td>
    <td>array of strings</td>
    <td>
        Full toolbox commands (with PIN/shortcut/PLT special directive) that will begin to run as soon as the service is
        online.
        <br/>
        Leave empty if you do not plan for any command to run automatically, you can still add transient commands using
        the HTML form.
    </td>
</tr>
</table>

Here is an example setup:
<pre>
{
    ...

    "HTTPHandlers": {
        ...

        "RecurringCommandEndpoint": "/very-secret-recurring-commands",
        "RecurringCommandEndpointConfig": {
            "RecurringCommands": {
                "my-secret-channel-alpha": {
                    "IntervalSec": 60,
                    "MaxResults": 10,
                    "PreConfiguredCommands": [
                        "VerySecretPassword.e info",
                        "VerySecretPassword.s date",
                    ]
                },
                "my-secret-channel-zulu": {
                    "IntervalSec": 120,
                    "MaxResults": 10,
                    "PreConfiguredCommands": [
                        "VerySecretPassword.il MyEmailInbox",
                    ]
                }
            }
        },


        ...
    },

    ...
}
</pre>

## Run
The form is hosted by web server, therefore remember to [run web server](https://github.com/HouzuoGuo/laitos/wiki/Daemon:-web-server#run).

## Usage
Pre-configured commands (if any) will begin to run automatically as soon as laitos starts up. To retrieve command
results, use an HTTP client (such as web browser) to access the endpoint URL (HTTP GET):

    /very-secret-recurring-commands?retrieve=my-secret-channel-alpha

The historic command results will be returned in a JSON array and then immediately deleted.

To add transient toolbox commands, clear transient toolbox commands, and to push messages directly into command results,
access the endpoint URL without additional parameter and use the HTML form:

    /very-secret-recurring-commands

## Tips
Make sure to choose a very secure URL for both the endpoint and channel names, and make sure they are only known by
designated users of this service!