  .org $0000

  .org $0200
  sec
  lda #$ff
  ldy #$0d
  sbc $02f0,y
  ldy #$0e
  sbc $02f0,y
  ldy #$0f
  sbc $02f0,y
  ldy #$10
  sbc $02f0,y
  ldy #$11
  sbc $02f0,y

  .org $02fd
  .byte $03
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
