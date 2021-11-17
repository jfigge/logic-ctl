  .org $0000

  .org $0200
  sec
  lda #$ff
  ldy #$0d
  sbc $02f0,y
  iny
  sbc $02f0,y
  iny
  sbc $02f0,y
  iny
  sbc $02f0,y
  iny
  sbc $02f0,y

  .org $02fd
  .byte $03
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
