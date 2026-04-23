# Feature "Commission artistique" — Vision

> **Purpose of this document:** This document expresses needs and constraints. It does not
> prescribe a solution, neither functional nor technical.

## Need

The association has members that are part of the "Commission artistique", which is where the program for next concerts are chosen.

As of now, everything is handled in a Google Sheet:
- A tab where the sheet music proposals are gathered:
  - Title
  - Composer (optional)
  - Musician proposing it
  - When (roughly: can be a year, or a year and a month)
  - One or several links to listen (usually Youtube) (can be empty)
  - One or several links to buy the sheet music (can be empty)
  - Duration (optional)
  - Difficulty, from 1 to 5 (optional)
  - A flag is the "Commission artistique" think it's good to play indoor
  - A flag is the "Commission artistique" think it's good to play outdoor
  - A column where CA-members comment, all in the same cell
- Actually, for traceability, one of the above tab is created at each new CA-meeting, removing the lines not needed anymore
- A tab for each CA-meeting, containing:
  - One column per concert for which we need to define a program
    - With the duration of the concert
    - And the number of rehearsals until the concert
  - One line for each CA-member, where people can suggest pieces based on the proposals, for each concert
  - One line giving the final decision of the CA
  - Finally, a line extracting the sheet musics to buy

The need is to integrate the existing process in the tool, so that we can decomission entirely the Google Sheet.

## Requirements

Add a role to flag "Commission artistique" members, opening new permissions in the webapp.

CA-members can add a sheet music proposal. Pieces of information:
  - Title
  - Composer (optional)
  - Musician proposing it (refering to an existing musician)
  - Date (default today)
  - One or several links to listen (can be empty)
  - One or several links to buy the sheet music (can be empty)
  - Duration (optional)
  - Difficulty, from 1 to 5 (optional)
  - Comment (text)

Each CA-member can modify a proposal.

Each CA-member can comment on proposals:
  - Give a note from 1 to 5 to say if they'd like to play it indoor
  - Give a note from 1 to 5 to say if they's like to play it outdoor
  - Write a comment
  
An event should now have a flag "is published", so that Admin can create it for CA-members, before it being visible for musicians. Concerts can have an expected duration, and a "Commission artistique comment".

CA-member can propose a program for a concert, picking sheet music proposals.
If the "sheet music management" is already implemented, CA-member could also propose an already owned piece.
CA-member can add a comment with his program proposal.

Any CA-member could complete the CA decision for the program. It has the same attributes as an individual proposal, but it's taken agreed by everyone in meeting.

If the "sheet music management" is already implemented, an Admin can later create a sheet music from a proposal, when it has been bought. Or link an existing one if it has been created independently.
