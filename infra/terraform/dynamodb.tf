resource "aws_dynamodb_table" "todo" {
  name         = "wakaba-todo"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "channel_id"

  attribute {
    name = "channel_id"
    type = "S"
  }

  tags = {
    Project = var.project_name
  }
}
