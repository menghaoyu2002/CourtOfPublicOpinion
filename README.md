# CourtOfPublicOpinion
general opinion from the internet 

1. i scrape youtube videos transcript
2. i process each word in transcript
3. if word is nice +1 if word is bad -1
4. if nice word before bad word -2 (e.g. "incredibly bad")
5. sum up values, average over number of words
6. get average sentiment per word of the video
7. sum up all scores from videos
8. get average percentage of good/bad words per video

bad things:
- could not get goroutines to work with opening tabs and youtube videos
- bad things happen if i try. tabs die or run out of memeory
- this means that we can only process one youtube video at a time
- this means that this api call is incredbly slow and blocks the main thread
- this means that this is rediculously slow as the number of users go up
- unscable
- on average takes 3 minutes to process one request
- figure out how to unblock the main thread maybe?
