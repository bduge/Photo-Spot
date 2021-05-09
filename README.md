# Photo Spot

**Backend developer challenge for Shopify internship application (Fall 2021) by Bill DuGe**


This is my personal take on the challenge to create an image repository. This is a web application that allows users to create their own photo contests with a specific description. Anyone can enter into the contest by submitting their own photo which matches the description/theme of the contest.

After the submission period is over, the contest opens up to users for voting. After the voting period ends, the winner will be announced and the winning submission can be viewed on the contest page.

---

### How to use

Requirements:
- Golang (https://golang.org/doc/install)
- MongoDB (https://docs.mongodb.com/manual/installation/)

Setup:

- Clone repository into a local directory
- Start `mongod` service in background (method depends on platform, refer to MongoDB documentation for detailed instructions)
- Run `go run .` to start server. Setup to run on `localhost:3000` by default. This can be changed at the bottom of `server.go`
- Run `go test` to execute unit tests
- Run `go mod download` to download dependencies if necessary

User Guide / Features:

- Create an account or login from home page
- Logged in users can view all contests, click on one to view more details
- Users can create their own contests, only the creator will be able to start/end the voting period for a contest
- If a contest is in its submission period, the user may select an image and make an entry into the contest. A user can only make 1 entry per contest.
- If a contest is in its voting period, its details page will display the submissions and the user may vote on their favourite. A user can only vote once.
- If a contest is concluded, the user will not be able to engage with the contest but can view the winner(s) from the voting results

---

### Possible improvements

Given that this project was created for the Shopify backend developer challenge as part of the intern application process, there was limited time to implement all the ideas I had. Here are some possible feautres for future development

- Allow users to searc contests by status (i.e. Open, Voting, or Concluded) or by name/keywords
- Add a user page which displays their previously submitted images and any contests they've won
- Add private contests for friends or communities
- Allow other media formats (i.e. Videos, GIFS, etc)
- Polish user interface + make responsive for mobile devices
- If publishing to production, implement security features like encrypting password encryption, secure password requirements, and verifying API requests
