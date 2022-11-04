package main

import (
	log "github.com/sirupsen/logrus"
)

type SceneNode struct {
	NodeId               CrdtId
	Name                 LwwString
	Visible              LwwBool
	AnchorId             LwwCrdt
	AnchorMode           LwwByte
	AnchorThreshold      LwwFloat
	AnchorInitialOriginX LwwFloat
	Bob                  []byte
}

func readSceneNode(decoder *ChunkDecoder) (node SceneNode, err error) {
	node.NodeId, _, err = decoder.ExtractCrdtId(1)
	if err != nil {
		return
	}
	node.Name, _, err = decoder.ExtractLwwString(2)
	if err != nil {
		return
	}

	node.Visible, _, err = decoder.ExtractLwwBool(3)
	if err != nil {
		return
	}

	selectedId, hasAnchor, err := decoder.ExtractCrdtId(4)
	if err != nil {
		return
	}
	if hasAnchor {
		node.AnchorId.Value = selectedId

		node.AnchorMode.Value, _, err = decoder.ExtractByte(5)
		if err != nil {
			return
		}

		node.AnchorThreshold.Value, _, err = decoder.ExtractFloat(6)
		if err != nil {
			return
		}

	} else {
		node.AnchorId, _, err = decoder.ExtractLwwCrdt(7)
		if err != nil {
			return
		}
		node.AnchorMode, _, err = decoder.ExtractLwwByte(8)
		if err != nil {
			return
		}

		node.AnchorThreshold, _, err = decoder.ExtractLwwFloat(9)
		if err != nil {
			return
		}

		node.AnchorInitialOriginX, _, err = decoder.ExtractLwwFloat(10)
		if err != nil {
			return
		}
	}
	node.Bob, err = decoder.ExtractBob()
	if err != nil {
		log.Warn("can't get bob", err)
		return
	}

	return
}
