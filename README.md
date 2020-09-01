# GoClue
This is a Google Drive Console client written in Golang during the COVID-19 pandemic, I have to stay at home...
It is based on google drive api v3. Currently only supports running in the Linux environment, because some internal Linux commands are used.
In the future, I will try to add support for the windows, but maybe there are not many people who like console programs under the windows platform.

Install

   1. Go to https://developers.google.com/drive/api/v3/enable-drive-api, 
Enable the Google Drive API
To interact with the Drive API, you need to enable the Drive API service for your app. You can do this in the Google API project for the app.
To enable the Drive API, complete these steps:

    1.1 Go to the Google API Console.https://console.developers.google.com/

    1.2 Select a project.

    1.3 In the sidebar on the left, expand APIs & auth and select APIs.

    1.4 In the displayed list of available APIs, click the Drive API link and click Enable API.

  2.  Create credentials to access your enabled APIs
    
    2.1 Reference https://developers.google.com/identity/protocols/oauth2/
    
  3.  Download from github
    
    3.1 run  go install -i -gcflags="-N -l" goclue.go
  
  4. Enjoy
    ![image](https://github.com/roleyzhang/GoClue/blob/master/img/pic-selected-200831-2255-59.png)
    ![image](https://github.com/roleyzhang/GoClue/blob/master/img/pic-selected-200831-2257-00.png)
    ![image](https://github.com/roleyzhang/GoClue/blob/master/img/pic-selected-200831-2300-53.png)
    ![image](https://github.com/roleyzhang/GoClue/blob/master/img/pic-selected-200831-2301-55.png)
    ![image](https://github.com/roleyzhang/GoClue/blob/master/img/pic-selected-200831-2251-18.png)

    
