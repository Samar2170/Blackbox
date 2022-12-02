### Upload
#### Small Files < 2GB
* Via Web Browser or via API

#### Multiple Small Files 
* Via Web Browser or via API

#### Big Files > 2GB
* Via Go Client. Using Chunking

#### Folders 
* Via Go Client. Using Compression & Chunking , combing and unziipping on the server. 




### Download
* Via Web Browser or via API

### API ENDPOINTS
* /upload - Done
* /download - Done
* /uploads - Done
* /bigupload
* /login - Done
* /logout 
* /register - Done
* /spaceused




#### Models
* User
* File


#### Process
* Login -> Verify User -> Provide token
* Upload -> Recieve File -> Get Metadata -> Store in DB -> Store in File System -> Return File Signed Url



#### Thumbnail Service

