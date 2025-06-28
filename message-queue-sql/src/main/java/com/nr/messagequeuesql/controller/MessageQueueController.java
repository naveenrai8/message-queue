package com.nr.messagequeuesql.controller;

import com.nr.messagequeuesql.dto.MessageRequestDto;
import com.nr.messagequeuesql.dto.MessageResponseDto;
import com.nr.messagequeuesql.service.MessageQueueService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/messages")
@RequiredArgsConstructor
@Slf4j
public class MessageQueueController {

    private final MessageQueueService service;

    @GetMapping
    public ResponseEntity<MessageResponseDto> getMessages(@RequestParam(required = false) Integer count) {
        log.info("Received request to get all messages");
        if (count == null) {
            count = 1;
        }

        return ResponseEntity.ok(this.service.getMessagesByCount(count));
    }

    @PostMapping
    public ResponseEntity<Void> addMessage(@RequestBody MessageRequestDto messageRequestDto) {
        this.service.addMessage(messageRequestDto);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("{id}")
    public ResponseEntity<Void> deleteMessage(@PathVariable String id) {
        this.service.deleteMessage(id);
        return ResponseEntity.ok().build();
    }
}
