resource "aws_iam_role" "lambda_exec" {
  name = "${var.project_name}-lambda-exec"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "dynamodb_access" {
  name = "dynamodb_access"
  role = aws_iam_role.lambda_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Effect   = "Allow"
        Resource = aws_dynamodb_table.todo.arn
      },
    ]
  })
}


# The ZIP file is expected to be built by the CI/CD pipeline or Makefile
variable "lambda_zip_path" {
  description = "Path to the lambda deployment package"
  type        = string
  default     = "../../apps/wakaba/dist/function.zip"
}

resource "aws_lambda_function" "app" {
  function_name = "${var.project_name}-function"
  
  filename         = var.lambda_zip_path
  source_code_hash = fileexists(var.lambda_zip_path) ? filebase64sha256(var.lambda_zip_path) : null
  
  handler = "bootstrap"
  runtime = "provided.al2"
  architectures = ["x86_64"]

  role = aws_iam_role.lambda_exec.arn

  environment {
    variables = {
      DISCORD_PUBLIC_KEY  = var.discord_public_key
      DISCORD_BOT_TOKEN   = var.discord_bot_token
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.todo.name
    }
  }
}

variable "discord_public_key" {
  description = "Discord Public Key"
  type        = string
  sensitive   = true
}

variable "discord_bot_token" {
  description = "Discord Bot Token"
  type        = string
  sensitive   = true
}
