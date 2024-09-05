# Google photos takeout organizer

I've recently exported all my photos from google photos, but there were duplicates and disogranization. this cli is my
personal solution to this problem.

## Goals

1. Remove duplication, only keep one version of the photo
2. Keep all metadata
3. Keep metadata about albums, but not as directory structures.

## TODO:

* [x] Travers tgz file and list all files.
* [x] Calculate hashes of archive entries.
* [x] Write out to an index file that can be used to analyze the data afterwards.
* [x] Add some cli and cleanup.
* [] Organize stuff out of main.go
* [] Move the initialization from `NewArchive` to the Entries, allows reusing the archive struct.
* [] Create a "global" progress tracker to report failed and skipped files instead of the current solution.
* [] Index all the existing archives.
* [] Analyze the data
    * [] How many photos are duplicated.
    * [] Is the file name good enough for deduplication? any entries with the same hash and different file names?
    * [] Estimate the size of the photos without duplication. Is it worth it to deduplicate ? I estimate a reduction of
      size to around 300 GB instead of 800 GB.
* [] Extract Album membership from the folder name.
* [] Extract to a flat structure.
* [] Try to keep the file creation timestamp intact.

