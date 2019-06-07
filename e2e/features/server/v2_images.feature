Feature: API Endpoint /v2/images/{reference}
   V2 Images API Endpoint
   Scenario: Discover an unknown image
      Given the Docker image "127.0.0.1:52854/golang:latest" pushed
      And imagespy API started
      When sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:latest"
      And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:latest"
      Then the API responds with status code "200"
      And the API responds with a body of "fixtures/results/a.json"

   Scenario: Discovering a newer version of an image sets the newer image as the latest image of the older version
      Given the Docker image "127.0.0.1:52854/golang:1.12.3" pushed
      And imagespy API started
      And sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:1.12.3"
      And the Docker image "127.0.0.1:52854/golang:1.12.4" pushed
      When sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
      And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:1.12.3"
      Then the API responds with status code "200"
      And the API responds with a body of "fixtures/results/b.json"

   Scenario: Discovering a newer version of an image sets the newer image as the latest image of itself
      Given the Docker image "127.0.0.1:52854/golang:1.12.3" pushed
      And imagespy API started
      And sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:1.12.3"
      And the Docker image "127.0.0.1:52854/golang:1.12.4" pushed
      When sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
      And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
      Then the API responds with status code "200"
      And the API responds with a body of "fixtures/results/c.json"

   Scenario: Discovering an image with the same digest adds the new tag
      Given the Docker image "127.0.0.1:52854/golang:1.12.4" pushed
      And the Docker image "127.0.0.1:52854/golang:latest" pushed
      And imagespy API started
      When sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
      And sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:latest"
      And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
      Then the API responds with status code "200"
      And the API responds with a body of "fixtures/results/d.json"

   Scenario: Discovering an image again does not work
      Given the Docker image "127.0.0.1:52854/golang:1.12.4" pushed
      And imagespy API started
      When sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
      And sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
      Then the API responds with status code "409"
