  .org $0000

  .org $0200
  ldx #$0e
  inc $02f0,x
  inx
  inc $02f0,x
  inx
  inc $02f0,x

  .org $02fe
  .byte $7f
  .byte $ff
  .byte $01
