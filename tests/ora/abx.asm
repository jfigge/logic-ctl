  .org $0000

  .org $0200
  lda #$00
  ldx #$0e
  ora $02f0,x
  ldx #$0f
  ora $02f0,x
  ldx #$10
  ora $02f0,x
  ldx #$11
  ora $02f0,x

  .org $02fe
  .byte $c0
  .byte $03
  .byte $3c
  .byte $f0
