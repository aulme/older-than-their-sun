# Older Than Their Sun

A generator of deep history for the Milky Way.

It lays out a real field of stars, runs tens of millions of years of rise and fall across them, and prints the result as legends you can read — including what each surviving people still believes about it, which is not the same thing as what happened.

A learning project in procedural world generation, in the spirit of Dwarf Fortress's legends mode and Caves of Qud: simulate the big picture first, generate the details on demand from that history. The goal for now is a good generator, not a game.

**For now this project is almost exclusively vibe coded. It's an exploration of procedural history generation in principle and how far it can go. **

## Run it

```sh
go run ./cmd/worldgen -seed 5 -stars 200 -legends   # generate a world, write out/5, print its legends
go run ./cmd/worldgen -read out/5                   # print the legends of a run written earlier
go run ./cmd/worldgen -read out/5 -full             # also every dead people's telling, and everyone's tech
go run ./cmd/worldgen -map                          # a chart of the galaxy and its named places
go run ./cmd/worldgen -at "Cygnus X-1" -stars 400   # anywhere: a preset, a named feature, or x,y,z in kpc
```

Go 1.27, no dependencies. A run takes a few seconds to a minute, depending on how long the age turns out to last. The same seed gives the same world byte for byte, and tests pin the digests of two of them.

## What comes out

Where you are, and what kind of galaxy it is at the moment:

```
seed 5: 200 stars within 150 ly of Sol
The place: the Sun's neighbourhood, in the Middle Disc on the inner edge of the Local Arm, 8.2 kpc from the centre.
  The stars stand about as they do around the Sun, 13 light years apart in this thinned field.
  This is the Local Arm: a short spur between the two great arms; the Sun rides its inner edge, with Orion and Cygnus for neighbours.
  111 of the stars are real, with the worlds Earth knows of; the rest are drawn to the laws of the place.
  ...
the cycle: period 1143 Myr, fade 22 Myr; the current age dawned 39.68 Myr ago, fertility now 16.0% of its dawn, next dawn in 1103 Myr
```

Then the ages that came before this one, of which nothing survives but ruins and rumour:

```
=== THE AGES OF MYTH (7.04 Gyr ago to 39.68 Myr ago) ===
  6.79 Gyr ago     The dawn of an age. Everywhere at once, things start to think.
  6.79 Gyr ago     Somewhere, minds that ran on the decay of heavy elements and were therefore very patient rises.
  6.79 Gyr ago     It leaves a corridor of darkness where no light crosses.
  6.79 Gyr ago     It ends.
  6.79 Gyr ago     Somewhere, a people who folded themselves into a smaller dimension to save on entropy rises.
  6.77 Gyr ago     Somewhere, something that used stars the way others use fire rises.
  6.77 Gyr ago     It is gone. Its works remain.
  ...
  6.71 Gyr ago     The age wanes. Nothing new rises, and what remains dwindles. What is left is swept up by a long silence, in which the remaining few forgot each other.
```

Then the age itself, some thousands of events of it, and what is still standing at the end:

```
  The Qi on Flegetonte, interstellar, holding a single world. a caste society, conquerors, cautious, short-lived. Now: military overwhelming, survival formidable, social overwhelming; 3 ships in one fleet, 5 guns over one world; settled; sick with the Grey Wasting these 3583 thousand years, carried and sick with the Certainty That Bends the Knee these 603 thousand years, carried.
```

And then the same history again, as its survivors hold it:

```
The Found Ones tell it so (9 things held, 7 of them myth, 41 forgotten, 11 retold):
  long ago           a nameless sickness came to us. It was the doing of the treacherous Dath.
  long ago           a nameless sickness came to us. The old songs are about it.
  long ago           In the age of heroes, we learned that the galaxy had done all this before, and would again.
  long ago           In the age of heroes, we gained directed evolution, a body for every world.
  long ago           the Endless Laugh came to us. It is not forgotten.
  about 2.90 Myr ago The monsters of the Dath made the Endless Laugh for us and hid it in what they sent. It is not forgotten.
  long ago           The eaters of worlds poisoned the Found Ones, and we paid for it. Every child knows it.
  long ago           We and the faithless ones made peace. That was before we knew them.
  2.55 Myr ago       the Wandering Prayer came to us. We had done nothing to deserve it.
```

Nine things held and forty-one forgotten, out of an age they lived through from end to end. *Long ago* is what a date looks like once a tale has worn to myth; *a nameless sickness* is a plague worn past its own name; *the treacherous Dath* is not what the Dath are, it is how this people speaks of them. Whom a tale blames is part of what is held, and it can be wrong.

## The telling

This is the part the rest is built around. **Nothing in the simulation decides using the world as it is.** A people decides using what it holds of the world, and what it holds decays.

A tale is exact, then worn, then myth. Tales are forgotten outright, told second hand with the teller's slant on them, blamed on the wrong people, and left behind as testaments to be found and retold long after the people that lived them is gone. A people's temperament, the count of its own griefs, its wisdom and its list of monsters are all read out of its telling, so two peoples that saw the same war go on to act on different wars.

## What it simulates

- **A real substrate.** 809 catalogue stars (HYG, and the NASA Exoplanet Archive for the planets), with the rest of a field drawn to the laws of wherever you point it: metallicity, crowding, hazard, the arms and the bar and the bulge.
- **A galaxy with a cycle.** Ages dawn everywhere at once, wane and fade. A run is one age, with the ages of myth behind it and everything they left lying around.
- **Peoples, not empires.** Eight temperament dials and 67 traits, a morality, a body and a biology, growing along a tech tree of 112 nodes in five eras. They meet, trade, teach, lie, make pacts, break them, conquer, uplift machines, sunder into heirs, ossify, and end — 262 kinds of event in all.
- **Plagues** with symptoms, courses and carriers, doctrines among them; **wars** fought with fleets and guns; **remains** left by the dead ages for the living to find and misunderstand.

## The run directory

A run writes one self-contained directory, and nothing in it is a sentence: ids, numbers, and keys that resolve in the lookups under `data/`. The `FORMAT.md` written into each run is the contract for every file and field, and a test fails when the writer emits a field the contract does not list. `internal/legends` is the only thing that turns any of it into English, which keeps the simulation clear of prose and the prose clear of decisions.

## Layout

| path | what |
|---|---|
| `cmd/worldgen` | the generator and the legends reader |
| `cmd/techstats` | batch runs and the measurement reports |
| `internal/history` | the simulation: peoples, decisions, the telling, war, plague, decline |
| `internal/galaxy` | the substrate and the star catalogue |
| `internal/legends` | everything that produces a sentence |
| `internal/record`, `internal/writer` | the run directory and its contract |
| `data/` | the lookups: tech, traits, events, plagues, names, epithets |
| `specs/`, `DESIGN_NOTES.md` | the design and the plans |
| `reports/tech/` | a ten-world batch and what it says about the tech tree |

## Reading further

`DESIGN_NOTES.md` is the canonical design document: what this is, why the galaxy works the way it does, and every decision worth arguing about. `specs/plan.md` is the open plan; `specs/done/plan.md` is the one that built the current simulation, with its ledger of what each step cost and what it measured. `reports/tech/report.md` is what ten worlds say about the tech tree.

## Status

Work in progress and changing weekly: no stable interfaces, no releases, and the output is settled only as far as `FORMAT.md` says it is. Written with the help of Claude Code, as the commit trailers record.

## Licence

The code is MIT; see `LICENSE`.

The star catalogue is not the project's to relicense. `internal/galaxy/catalog_data.go` is generated by `cmd/mkcatalog` from the [HYG database](https://github.com/astronexus/HYG-Database), which is **CC BY-SA 4.0**, and from the [NASA Exoplanet Archive](https://exoplanetarchive.ipac.caltech.edu/) for the planets, which is public domain with a request for acknowledgement. Reuse that file under those terms, and keep the attribution its header carries.
