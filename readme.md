# Google photos takeout organizer 

I've recently exported all my photos from google photos, but there were duplicates and disogranization. this cli is my personal solution to this problem.


## Goals
1. Remove duplication, only keep one version of the photo
2. Keep all metadata
3. Keep metadata about albums, but not as directory structures.


## Tasks
### Find duplicates
I'm thinking all files in the archives and create a name, hash listing of all, once done it would be trivial to find out duplicates from those.

1. Traverse a tgz file, list all files there
2. Calculate a hash
3. dump the output line by line

