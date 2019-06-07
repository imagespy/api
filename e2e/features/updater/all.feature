Feature: Command "updater all"

  Scenario: Find the new version of an image
    Given the Docker image "127.0.0.1:52854/golang:1.12.3" pushed
    And imagespy API started
    And sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:1.12.3"
    And the Docker image "127.0.0.1:52854/golang:1.12.4" pushed
    When running the updater comand "all"
    And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
    Then the API responds with status code "200"
    And the API responds with a body of "fixtures/results/c.json"
