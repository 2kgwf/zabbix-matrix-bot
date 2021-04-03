# Zabbix Matrix bot

A simple [Matrix](https://matrix.org) bot for Zabbix

TODO: improve documentatio and cleanup

### Setup

- Create a Matrix user for the bot and acquire the access token for it
- Set up the bot (Docker preferred) with the environment variables defined in `docker-compose.yml` (note: the bot will accept room invites only from the defined admin user)
- Create a Zabbix user for the bot and grant it read-only access (the bot command `!problems` will list all active problems on hosts it has read access to)
- Copy the `notify_matrix.sh` to `/usr/lib/zabbix/alertscripts/notify_matrix.sh` on the Zabbix server and adjust the host on the curl command, if needed
- Create a media type in Zabbix (Adminstration -> Media types):
  - Name: Matrix
  - Type: Script
  - Script name: notify_matrix.sh
  - Script parameters:
    - `{ALERT.SENDTO}`
    - `{ALERT.SUBJECT}`
    - `{ALERT.MESSAGE}`
  - On the Message templates tab you can and should adjust the templates. You can use emojis and "matrix html" in formatting. Note: the Subject will be used for "matching" related updates and it will not be included in the message. For a notification with the same subject, the bot will edit the previous message with the same subject rather than posting a new message, this allows problem messages to be edited in-place to state "Resolved" when they get resolved, for example. Easiest way to achieve this is to just include the `{EVENT.ID}` only in the subject.
- Create Media(s) for user(s) you wish to get the alerts for, ie. the user for the bot or your own user (Administration -> Users, pick an user and Add on the Media tab):
  - Type: Matrix
  - Send to: <the matrix ID of the room to send to, ie. `!RsdgHGdfgGREHewg:example.com`>
  - Rest of the options based on your preference

### Example Message templates

- Message type: Problem
- Subject: `Problem: {EVENT.ID}`
- Message:

```
<h4>⚠️ Problem: {EVENT.NAME}</h4> Problem started at <b>{EVENT.TIME}</b> on <b>{EVENT.DATE}</b>
Host: <b>{HOST.NAME}</b>
Severity: <b>{EVENT.SEVERITY}</b>
{TRIGGER.URL}
Original problem ID: <i>{EVENT.ID}</i>
<i>{TRIGGER.DESCRIPTION}</i>
```

- Message type: Problem recovery
- Subject: `Problem: {EVENT.ID}`
- Message:

```
<h4>✅ Resolved: {EVENT.NAME}</h4> Resolved at {EVENT.RECOVERY.TIME} on {EVENT.RECOVERY.DATE}. Problem started at <b>{EVENT.TIME}</b> on <b>{EVENT.DATE}</b>
Host: <b>{HOST.NAME}</b>
Severity: <b>{EVENT.SEVERITY}</b>
{TRIGGER.URL}
Original problem ID: <i>{EVENT.ID}</i>
<i>{TRIGGER.DESCRIPTION}</i>
```

- Message type: Problem update
- Subject: `Problem: {EVENT.ID}`
- Message:

```
<h4>⚠️ Problem: {EVENT.NAME}</h4> Problem started at <b>{EVENT.TIME}</b> on <b>{EVENT.DATE}</b>
Host: <b>{HOST.NAME}</b>
Severity: <b>{EVENT.SEVERITY}</b>
{TRIGGER.URL}
Original problem ID: <i>{EVENT.ID}</i>
<i>{TRIGGER.DESCRIPTION}</i>

{USER.FULLNAME} {EVENT.UPDATE.ACTION} problem at {EVENT.UPDATE.DATE} {EVENT.UPDATE.TIME}.
{EVENT.UPDATE.MESSAGE}

Current problem status is {EVENT.STATUS}, acknowledged: {EVENT.ACK.STATUS}.
```
