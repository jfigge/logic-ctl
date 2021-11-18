  .org $0000

  .org $0200
  ldx #$0e
  dec $02f0,x
  inx
  dec $02f0,x
  inx
  dec $02f0,x

  .org $02fe
  .byte $ff
  .byte $01
  .byte $02
