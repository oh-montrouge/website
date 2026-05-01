# Sheet music management — Vision

> **Purpose of this document:** This document expresses needs and constraints. It does not
> prescribe a solution, neither functional nor technical.

## Need

Currently, all the orchestra sheet musics are stored in the Google Drive.
In the V1, the new website only integrated a link to the Google Drive root folder.

We want to integrate the sheet musics in the web application, so that musicians can easily search for them.

## Requirements

All stored under GDrive.
We can reference from the DB.
And have a search engine.
And put the sheet music links in the events.
And show the RSVP instrument or main instrument in priority.

We should replace the link to the Google Drive folder in the navigation bar, so that it redirects to a new page.

This page should allow to search among the sheet musics. 
Clicking on a sheet music, a musician see the details, and can download the individual parts.

Admins can add a new sheet music:
- From a propal (link with commission-artistique.md)
- From a "New" button

A sheet music should have the following pieces of information:
  - Title
  - Composer (optional)
  - Arranger (optional)
  - Duration (optional)
  - Acquisition date (optional)
  - Difficulty, from 1 to 5 (optional)
  - Comment (text)
  - One or several links to listen (can be empty)
  - A list of tags, could be "Classique", "Jeux-vidéo", "Film", "Harmonie", "Chanson populaire", "Disney", "Cérémonies patriotiques", "Banda". Can be empty.

To a sheet music can be associated several files (individual parts).
A part is associated to an instrument. However, there can be several parts per instrument.

The association can be created manually, specifying a label and a link.

However, we could have a feature to give a Google Drive folder link, containing PDF files (should ignore what is not PDF).
The tool can create one association for each file. The label could be derived from the file name, and the instrument should theoretically be in the file name.

Using the same mechanism, we will want to backfill the data crawling the root Google Drive folder.

From the next events list, a Musician should see the parts for the instrument he has declared he will play, or his main instrument. A link should allow to access the sheet music details with all parts?
