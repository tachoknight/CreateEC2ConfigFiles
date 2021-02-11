# Create EC2 Config Files #
This is a simple program to generate ~/.ssh/config files for EC2 instances. 

## Usage ##

The basic usage is:

`go run main.go -key=_<EC2 tag key>_ -value=_<EC2 tag value>_ -region=_<whatever region your instances are located>_ > _<config file>_`

so for example:

`go run main.go -key=Environment -value=production -region=us-west-1 > ec2_production`

will produce a `ec2_production` file that you can put in your `~/.ssh` directory and add it to your regular `config` file by adding a directive like:

`Include ec2_production`

### AWS Credentials ###
This program uses the AWS credentials, via the AWS SDK, stored either as environment variables or in the `~/.aws/credentials` file. 

## Notes ##
* This program assumes that an EC2 instance with a public ip is a "bastion" machine, and that all the instances with only a private IP will use the bastion machine as the proxy.
* It's assumed you have the proper keys to log into the instance; it does not read or do anything with keys other than indicate which key goes with which instance.
* EC2 instances can have the same name (e.g. `webserver`). This program will append a number to subsequent names automatically (e.g. `webserver-2`, `webserver-3`, etc.)
* The program writes to `stdout`, you must pipe it to a file manually
* If you want to change what the program writes, look for `Now let's put together the config file. ` around line 98 and you'll see the code that it uses to print the output. Feel free to make whatever changes are necessary for your environment.


Hopefully this program will be useful for somebody. Enjoy!