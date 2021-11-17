  .org $0000

  .org $0200
  sec
  lda #$ff
  ldx #$0d
  sbc $02f0,x
  inx
  sbc $02f0,x
  inx
  sbc $02f0,x
  inx
  sbc $02f0,x
  inx
  sbc $02f0,x

  .org $02fd
  .byte $03
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
