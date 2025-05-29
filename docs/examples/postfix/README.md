# Example Postfix MTA configuration

This set of files is an example of Postfix MTA configuration for Ovoo.

To support proper DKIM signature validation for incoming mail and proper signin for outgoing mail,
the [Postfix multi-instance](https://www.postfix.org/MULTI_INSTANCE_README.html) approach is [utilized](./main.cf).

The **[Postfix-In](./in)** instance is responsible for all incoming mail addressed to the domain Ovoo is serving.
This instance will check for SPF records and DKIM (using OpenDKIM) signatures of all the incoming emails and then pass it to the Ovoo Milter for validation and address rewrite.

In case the targeted Alias was found by Ovoo Milter and addresses has been rewritten, then the mail is passed to the **[Postfix-Out](./out)** instance (listening on 127.0.0.1:10026), which runs it through OpenDKIM again to sign the mail envelope with correct DKIM of the Ovoo domain.

<p align="left">
    <img width="100%" src="./../../assets/overview/ovoo_overview.svg" alt="Ovoo overview diagram" />
</p>
