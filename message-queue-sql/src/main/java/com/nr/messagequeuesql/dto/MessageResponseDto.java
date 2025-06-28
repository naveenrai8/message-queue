package com.nr.messagequeuesql.dto;

import lombok.Builder;

import java.util.List;

@Builder
public record MessageResponseDto (
        List<String> messages
){
}
