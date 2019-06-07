Feature: API Endpoint /v2/images/{reference}/layers
   V2 Image Layers API Endpoint
   Scenario: Layers of image which has no parents
      Given the Docker image "127.0.0.1:52854/golang:latest" pushed
      And imagespy API started
      When sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:latest"
      And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:latest/layers"
      Then the API responds with status code "200"
      And the API responds with a body of "fixtures/results/e.json"

   Scenario: Layers of image which has a parent
      Given the Docker image "127.0.0.1:52854/golang:latest" pushed
      And the Docker image "127.0.0.1:52854/debian:stretch-20190326" pushed
      And imagespy API started
      When sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:latest"
      And sending a "POST" request to "/v2/images/127.0.0.1:52854/debian:stretch-20190326"
      And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:latest/layers"
      Then the API responds with status code "200"
      And the API responds with a body of "fixtures/results/f.json"
