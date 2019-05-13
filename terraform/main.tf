# vars
variable "region" {
  type = "string"
  default = "us-west-1"
}

# provider
provider "aws" {
  profile = "jds"
  region     = "${var.region}"
}

# import
module "stinkyfingers" {
  source = "../../../../../../infrastructure/stinkyfingers"
}

# Lambda
resource "aws_lambda_permission" "jmeme" {
  statement_id  = "AllowExecutionFromApplicationLoadBalancer"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.jmeme.arn}"
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn = "${aws_lb_target_group.jmeme.arn}"
}

resource "aws_lambda_permission" "jmeme_live" {
  statement_id  = "AllowExecutionFromApplicationLoadBalancer"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_alias.jmeme_live.arn}"
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn = "${aws_lb_target_group.jmeme.arn}"
}

resource "aws_lambda_alias" "jmeme_live" {
  name             = "live"
  description      = "set a live alias"
  function_name    = "${aws_lambda_function.jmeme.arn}"
  function_version = "${aws_lambda_function.jmeme.version}"
}

resource "aws_lambda_function" "jmeme" {
  filename         = "../jmeme.zip"
  function_name    = "jmeme"
  role             = "${aws_iam_role.lambda_role.arn}"
  handler          = "jmeme"
  runtime          = "go1.x"
  source_code_hash = "${filebase64sha256("../jmeme.zip")}"
  environment {
    variables = {
      VERIFICATION_TOKEN  = "${data.aws_ssm_parameter.verification_token.value}"
      AUTH_TOKEN          = "${data.aws_ssm_parameter.auth_token.value}"
      GOOGLE_API_KEY      = "${data.aws_ssm_parameter.google_api_key.value}"
      SLACK_HOOK_URL      = "${data.aws_ssm_parameter.slack_hook_url.value}"
    }
  }
}

data "aws_ssm_parameter" "verification_token" {
  name = "/jmeme/VERIFICATION_TOKEN"
}
data "aws_ssm_parameter" "auth_token" {
  name = "/jmeme/AUTH_TOKEN"
}
data "aws_ssm_parameter" "google_api_key" {
  name = "/jmeme/GOOGLE_API_KEY"
}
data "aws_ssm_parameter" "slack_hook_url" {
  name = "/jmeme/SLACK_HOOK_URL"
}


# IAM
resource "aws_iam_role" "lambda_role" {
  name = "jmeme-lambda-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "lambda_logging" {
  name = "lambda_logging"
  description = "IAM policy for logging from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role = "${aws_iam_role.lambda_role.name}"
  policy_arn = "${aws_iam_policy.lambda_logging.arn}"
}

# ALB
resource "aws_lb_target_group" "jmeme" {
  name        = "jmeme"
  target_type = "lambda"
}

resource "aws_lb_target_group_attachment" "jmeme" {
  target_group_arn  = "${aws_lb_target_group.jmeme.arn}"
  target_id         = "${aws_lambda_alias.jmeme_live.arn}"
  depends_on        = ["aws_lambda_permission.jmeme_live"]
}

resource "aws_lb_listener_rule" "jmeme" {
listener_arn = "${module.stinkyfingers.stinkyfingers_http_listener}"
priority = 21
  action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.jmeme.arn}"
  }
  condition {
    field = "path-pattern"
    values = ["/jmeme/*"]
  }
  depends_on = ["aws_lb_target_group.jmeme"]
}
