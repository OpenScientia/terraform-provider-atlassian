resource "atlassian_jira_issue_field_configuration" "example" {
  name = "foo"
}

resource "atlassian_jira_issue_field_configuration_item" "example_required" {
  issue_field_configuration = atlassian_jira_issue_field_configuration.example.id
  item = {
    id          = "customfield_10000"
    is_required = true
  }
}

resource "atlassian_jira_issue_field_configuration_item" "example_optional" {
  issue_field_configuration = atlassian_jira_issue_field_configuration.example.id
  item = {
    id          = "customfield_10001"
    is_required = false
  }
}

