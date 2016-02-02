// testing a deep diff
level1.level2.level3:
< here!
---
> over there!
- level1-1
- level1-3.level2-1
+ level1-3.level2-2
+ level1-2
===
===
{
  "level1": {
    "level2": {
	  "level3": "here!"
	}
  },
  "level1-1": {
  	"level2-1": false
  },
  "level1-3": {
  	"level2-1": false
  }
}
===
{
  "level1": {
    "level2": {
	  "level3": "over there!"
	}
  },
  "level1-2": {
  	"level2-1": false
  },
  "level1-3": {
  	"level2-2": false
  }
}
