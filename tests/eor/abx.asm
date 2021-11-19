  .org $0000

  .org $0200
  lda #$ff
  ldx #$0e
  eor $02f0,x
  inx
  eor $02f0,x
  inx
  eor $02f0,x
  inx
  eor $02f0,x

  .org $02fe
  .byte $55
  .byte $50
  .byte $05
  .byte $ff
